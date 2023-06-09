package db

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"fulfillmentd/server/db"
	"fulfillmentd/utils"
	"github.com/eluv-io/errors-go"
	elog "github.com/eluv-io/log-go"
	"github.com/jackc/pgx"
	"io/fs"
	"regexp"
	"strconv"
	"text/template"
	"time"
)

//go:embed sql/*.tmpl
var statementsFS embed.FS

var log = elog.Get("/fs/db")

type FulfillmentPersistence struct {
	pool            *db.ConnectionManager
	ethUrlByNetwork map[string]string
}

type SetupData struct {
	ContractAddress string   `json:"contract_address"`
	OfferId         string   `json:"offer_id"`
	Url             string   `json:"url"`
	Codes           []string `json:"codes"`
}

type RedemptionTransaction struct {
	ContractAddress string `json:"contract_address"`
	RedeemerAddress string `json:"user_address"`
	TokenId         int64  `json:"token_id"`
	OfferId         uint8  `json:"offer_id"`
	IsPending       bool   `json:"-"`
}

type FulfillmentRequest struct {
	Transaction string `json:"transaction"`
	UserAddress string `json:"user_address"`
	Network     string `json:"network"`
}

type FulfillmentResponse struct {
	ContractAddr string    `json:"contract_address"`
	OfferId      string    `json:"offer_id"`
	TokenId      string    `json:"token_id"`
	Claimed      bool      `json:"claimed"`
	UserAddr     string    `json:"user_address"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`

	Url  string `json:"url"`
	Code string `json:"code"`
}

func NewFulfillmentPersistence(cm *db.ConnectionManager, ethUrls map[string]string) *FulfillmentPersistence {
	log.Info("init FulfillmentPersistence", "cm", cm)
	return &FulfillmentPersistence{pool: cm, ethUrlByNetwork: ethUrls}
}

func (fp *FulfillmentPersistence) AvailableNetworks() (nets []string) {
	nets = utils.Keys(fp.ethUrlByNetwork)
	return
}
func (fp *FulfillmentPersistence) SetupFulfillment(setup SetupData) (err error) {
	log.Debug("SetupFulfillment", "setup", setup)
	if setup.ContractAddress == "" || setup.OfferId == "" || setup.Url == "" || setup.Codes == nil || len(setup.Codes) == 0 {
		log.Debug("invalid setup", "setup", setup)
		err = errors.NoTrace("invalid load setup", errors.K.Invalid, "setup", setup)
		return
	}

	for _, code := range setup.Codes {
		var stmt string
		if stmt, err = mergeTemplate("sql/add-mapping.tmpl", fp.context()); err != nil {
			return
		}

		var args []interface{}
		args = append(args, setup.ContractAddress)
		args = append(args, setup.OfferId)
		args = append(args, setup.Url)
		args = append(args, code)

		if _, err = fp.conn().Exec(stmt, args...); err != nil {
			return
		}
	}

	return
}

func (fp *FulfillmentPersistence) FulfillRedeemableOffer(request FulfillmentRequest) (resp FulfillmentResponse, err error) {

	var tx RedemptionTransaction
	if tx, err = fp.resolveTransaction(request); err != nil {
		log.Warn("error resolving tx", "error", err)
		err = errors.NoTrace("error resolving tx", errors.K.Invalid, "error", err, "request", request)
		return
	}
	offerId := fmt.Sprintf("%d", tx.OfferId)
	tokenId := fmt.Sprintf("%d", tx.TokenId)
	log.Debug("FulfillRedeemableOffer", "request", fmt.Sprintf("%+v", request), "tx", fmt.Sprintf("%+v", tx), "offerId", offerId, "tokenId", tokenId)

	if request.UserAddress != tx.RedeemerAddress {
		err = errors.NoTrace("mismatched user address", errors.K.Invalid, "request", request, "tx", tx)
		return
	}

	resp, err = fp.GetRedeemedOffer(tx.ContractAddress, offerId, tokenId)
	if err != nil {
		return
	}
	if resp.Claimed {
		err = errors.NoTrace("token already claimed", errors.K.Invalid, "request", request, "tx", tx)
		return
	}

	var stmt string
	templateArgs := fp.context()
	if stmt, err = mergeTemplate("sql/update-mapping.tmpl", templateArgs); err != nil {
		return
	}

	var args []interface{}
	args = append(args, tokenId)
	args = append(args, tx.RedeemerAddress)
	args = append(args, tx.ContractAddress)
	args = append(args, offerId)
	//log.Trace("FulfillRedeemableOffer", "stmt", stmt, "args", args)

	var rows *pgx.Rows
	if rows, err = fp.conn().Query(stmt, args...); err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		// fulfillment successful
		resp, err = scanFulfillmentData(rows, tx.ContractAddress, offerId, tokenId)
		if resp.Claimed {
			resp.UserAddr = tx.RedeemerAddress

			err = fp.markUrlAndCodeClaimed(resp.Url, resp.Code)
			if err != nil {
				return
			}
		}
	} else {
		// fulfillment failed; see why
		var unclaimed []string
		unclaimed, err = fp.GetUnclaimed(tx.ContractAddress, offerId)
		if err != nil {
			return
		}

		if len(unclaimed) == 0 {
			err = errors.NoTrace("no more redemption codes available", errors.K.NotFound, "request", request, "tx", tx)
		} else {
			err = errors.NoTrace("unable to redeem", errors.K.Invalid, "request", request, "tx", tx)
		}
	}

	return
}

// markUrlAndCodeClaimed marks the url and code as claimed in all other contracts, in case there are dups
func (fp *FulfillmentPersistence) markUrlAndCodeClaimed(url, code string) (err error) {
	var stmt string
	templateArgs := fp.context()
	if stmt, err = mergeTemplate("sql/mark-url-and-code-claimed.tmpl", templateArgs); err != nil {
		return
	}

	var args []interface{}
	args = append(args, url)
	args = append(args, code)

	var rows *pgx.Rows
	if rows, err = fp.conn().Query(stmt, args...); err != nil {
		return
	}
	defer rows.Close()

	var otherContracts []string
	otherContracts, err = scanDups(rows)
	if err != nil {
		return
	}
	if len(otherContracts) > 1 {
		log.Debug("marked this Url and Code claimed", "otherContracts", otherContracts)
	}

	return
}

func (fp *FulfillmentPersistence) GetRedeemedOffer(contractAddr, redeemableId, tokenId string) (resp FulfillmentResponse, err error) {
	var stmt string
	templateArgs := fp.context()
	if stmt, err = mergeTemplate("sql/get-mapping.tmpl", templateArgs); err != nil {
		return
	}

	var args []interface{}
	args = append(args, contractAddr)
	args = append(args, redeemableId)
	args = append(args, tokenId)
	//log.Trace("GetRedeemedOffer", "stmt", stmt, "args", args)

	var rows *pgx.Rows
	if rows, err = fp.conn().Query(stmt, args...); err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		resp, err = scanFulfillmentData(rows, contractAddr, redeemableId, tokenId)
	}

	return
}

func (fp *FulfillmentPersistence) GetUnclaimed(contractAddr, redeemableId string) (unclaimed []string, err error) {
	log.Debug("GetUnclaimed", "contractAddr", contractAddr, "redeemableId", redeemableId)
	var stmt string
	templateArgs := fp.context()
	if stmt, err = mergeTemplate("sql/get-unclaimed.tmpl", templateArgs); err != nil {
		return
	}

	var args []interface{}
	args = append(args, contractAddr)
	args = append(args, redeemableId)

	var rows *pgx.Rows
	if rows, err = fp.conn().Query(stmt, args...); err != nil {
		return
	}
	defer rows.Close()

	unclaimed = make([]string, 0)
	for rows.Next() {
		var url, code sql.NullString
		if err = rows.Scan(&url, &code); err != nil {
			return
		}
		if code.Valid {
			unclaimed = append(unclaimed, code.String)
		}
	}

	return
}

func scanFulfillmentData(rows *pgx.Rows, contractAddr, redeemableId, tokenId string) (row FulfillmentResponse, err error) {
	var claimed sql.NullBool
	var addr, url, code sql.NullString
	var created, updated sql.NullTime
	if err = rows.Scan(&claimed, &addr, &url, &code, &created, &updated); err != nil {
		return
	}
	if claimed.Valid {
		row = FulfillmentResponse{
			Claimed:  claimed.Bool,
			UserAddr: addr.String,
			Created:  created.Time,
			Updated:  updated.Time,
			Url:      url.String,
			Code:     code.String,

			ContractAddr: contractAddr,
			OfferId:      redeemableId,
			TokenId:      tokenId,
		}
	}

	return
}

func scanDups(rows *pgx.Rows) (otherContracts []string, err error) {
	otherContracts = make([]string, 0)

	for rows.Next() {
		var addr sql.NullString
		if err = rows.Scan(&addr); err != nil {
			return
		}
		if addr.Valid {
			otherContracts = append(otherContracts, addr.String)
		}
	}

	return
}

func (fd *FulfillmentResponse) ToTransaction() RedemptionTransaction {
	tid, _ := strconv.ParseInt(fd.TokenId, 10, 64)
	var oid uint8
	_, _ = fmt.Scan(fd.OfferId, &oid)
	return RedemptionTransaction{
		RedeemerAddress: fd.UserAddr,
		ContractAddress: fd.ContractAddr,
		TokenId:         tid,
		OfferId:         oid,
	}
}

var whitespace = regexp.MustCompile(`\s+`)

func mergeTemplate(path string, ctx map[string]interface{}) (stmt string, err error) {
	var b []byte
	if b, err = fs.ReadFile(statementsFS, path); err != nil {
		return
	}

	var t *template.Template
	if t, err = template.New(path).Parse(string(b)); err != nil {
		return
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, ctx); err != nil {
		return
	}

	stmt = whitespace.ReplaceAllString(buf.String(), " ")
	return
}

func (fp *FulfillmentPersistence) conn() *pgx.ConnPool {
	return fp.pool.GetConn()
}

func (fp *FulfillmentPersistence) context() map[string]interface{} {
	return map[string]interface{}{
		"database": "fulfillmentservice",
	}
}
