package db

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"fulfillmentd/server/db"
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
	pool   *db.ConnectionManager
	ethUrl string
}

type SetupData struct {
	ContractAddr string   `json:"contract_addr"`
	RedeemableId string   `json:"redeemable_id"`
	Url          string   `json:"url"`
	Codes        []string `json:"codes"`
}

type RedemptionTransaction struct {
	ContractAddress      string `json:"contract_addr"`
	RedeemerAddress      string `json:"user_addr"`
	TokenId              int64  `json:"token_id"`
	OfferId              uint8  `json:"offer_id"`
	TransactionIsPending bool   `json:"is_pending"`
}

type FulfillmentRequest struct {
	Transaction string `json:"transaction"`
	UserAddr    string `json:"user_addr"`
}

type FulfillmentData struct {
	ContractAddr string    `json:"contract_addr"`
	OfferId      string    `json:"offer_id"`
	TokenId      string    `json:"Token_id"`
	Claimed      bool      `json:"claimed"`
	UserAddr     string    `json:"user_addr"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`

	Url  string `json:"url"`
	Code string `json:"code"`
}

func (fd *FulfillmentData) ToTransactionData() RedemptionTransaction {
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

func NewFulfillmentPersistence(cm *db.ConnectionManager, ethUrl string) *FulfillmentPersistence {
	log.Info("init FulfillmentPersistence", "cm", cm)
	return &FulfillmentPersistence{pool: cm, ethUrl: ethUrl}
}

func (fp *FulfillmentPersistence) conn() *pgx.ConnPool {
	return fp.pool.GetConn()
}

func (fp *FulfillmentPersistence) context() map[string]interface{} {
	return map[string]interface{}{
		"database": "fulfillmentservice",
	}
}

func (fp *FulfillmentPersistence) SetupFulfillment(data SetupData) (err error) {
	log.Debug("SetupFulfillment", "data", data)
	if data.ContractAddr == "" || data.RedeemableId == "" || data.Url == "" || data.Codes == nil || len(data.Codes) == 0 {
		log.Debug("invalid data", "data", data)
		err = errors.NoTrace("invalid load data", errors.K.Invalid, "data", data)
		return
	}

	for _, code := range data.Codes {
		var stmt string
		if stmt, err = mergeTemplate("sql/add-mapping.tmpl", fp.context()); err != nil {
			return
		}

		var args []interface{}
		args = append(args, data.ContractAddr)
		args = append(args, data.RedeemableId)
		args = append(args, data.Url)
		args = append(args, code)

		if _, err = fp.conn().Exec(stmt, args...); err != nil {
			return
		}
	}

	return
}

func (fp *FulfillmentPersistence) FulfillRedeemableOffer(request FulfillmentRequest) (resp FulfillmentData, err error) {

	var tx RedemptionTransaction
	if tx, err = fp.resolveTransactionData(request); err != nil {
		log.Warn("error resolving tx", "error", err)
		return
	}
	offerId := fmt.Sprintf("%d", tx.OfferId)
	tokenId := fmt.Sprintf("%d", tx.TokenId)
	log.Debug("FulfillRedeemableOffer", "request", fmt.Sprintf("%+v", request), "tx", fmt.Sprintf("%+v", tx), "offerId", offerId, "tokenId", tokenId)

	if request.UserAddr != tx.RedeemerAddress {
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
		resp, err = scanFulfillmentData(rows, tx.ContractAddress, offerId, tokenId)
		if resp.Claimed {
			resp.UserAddr = tx.RedeemerAddress
		}
	} else {
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

func (fp *FulfillmentPersistence) GetRedeemedOffer(contractAddr, redeemableId, tokenId string) (resp FulfillmentData, err error) {
	var stmt string
	templateArgs := fp.context()
	if stmt, err = mergeTemplate("sql/get-mapping.tmpl", templateArgs); err != nil {
		return
	}

	var args []interface{}
	args = append(args, contractAddr)
	args = append(args, redeemableId)
	args = append(args, tokenId)

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

func scanFulfillmentData(rows *pgx.Rows, contractAddr, redeemableId, tokenId string) (row FulfillmentData, err error) {
	var claimed sql.NullBool
	var addr, url, code sql.NullString
	var created, updated sql.NullTime
	if err = rows.Scan(&claimed, &addr, &url, &code, &created, &updated); err != nil {
		return
	}
	if claimed.Valid {
		row = FulfillmentData{
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
