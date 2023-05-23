package db

import (
	"bytes"
	"database/sql"
	"embed"
	"fulfillmentd/authority/db"
	"github.com/eluv-io/errors-go"
	elog "github.com/eluv-io/log-go"
	"github.com/jackc/pgx"
	"io/fs"
	"regexp"
	"text/template"
	"time"
)

//go:embed sql/*.tmpl
var statementsFS embed.FS

var log = elog.Get("/fs/db")

type FulfillmentPersistence struct {
	pool *db.ConnectionManager
}

type SetupData struct {
	ContractAddr string   `json:"contract_addr"`
	RedeemableId string   `json:"redeemable_id"`
	Url          string   `json:"url"`
	Codes        []string `json:"codes"`
}

type TransactionData struct {
	UserAddr     string `json:"user_addr"`
	ContractAddr string `json:"contract_addr"`
	RedeemableId string `json:"redeemable_id"`
	TokenId      string `json:"token_id"`
}

type FulfillmentRequest struct {
	Transaction string `json:"transaction"`
	UserAddr    string `json:"user_addr"`
}

type FulfillmentData struct {
	ContractAddr string    `json:"contract_addr"`
	RedeemableId string    `json:"redeemable_id"`
	TokenId      string    `json:"Token_id"`
	Claimed      bool      `json:"claimed"`
	UserAddr     string    `json:"user_addr"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`

	Url  string `json:"url"`
	Code string `json:"code"`
}

func NewFulfillmentPersistence(cm *db.ConnectionManager) *FulfillmentPersistence {
	log.Info("init FulfillmentPersistence", "cm", cm)
	return &FulfillmentPersistence{pool: cm}
}

func (fs *FulfillmentPersistence) conn() *pgx.ConnPool {
	return fs.pool.GetConn()
}

func (fs *FulfillmentPersistence) context() map[string]interface{} {
	return map[string]interface{}{
		"database": "fulfillmentservice",
	}
}

func (fs *FulfillmentPersistence) SetupFulfillment(data SetupData) (err error) {
	if data.ContractAddr == "" || data.RedeemableId == "" || data.Url == "" || data.Codes == nil || len(data.Codes) == 0 {
		log.Debug("invalid data", "data", data)
		err = errors.NoTrace("invalid load data", errors.K.Invalid, "data", data)
		return
	}

	for _, code := range data.Codes {
		var stmt string
		if stmt, err = mergeTemplate("sql/add-mapping.tmpl", fs.context()); err != nil {
			return
		}

		var args []interface{}
		args = append(args, data.ContractAddr)
		args = append(args, data.RedeemableId)
		args = append(args, data.Url)
		args = append(args, code)

		if _, err = fs.conn().Exec(stmt, args...); err != nil {
			return
		}
	}

	return
}

func (fs *FulfillmentPersistence) ResolveTransactionData(request FulfillmentRequest) (data TransactionData, err error) {
	// TODO: figure out the data in the transaction
	data = TransactionData{}
	return
}

func (fs *FulfillmentPersistence) FulfillRedeemableOffer(request FulfillmentRequest) (resp FulfillmentData, err error) {
	var tx TransactionData
	if tx, err = fs.ResolveTransactionData(request); err != nil {
		return
	}

	if request.UserAddr != tx.UserAddr {
		err = errors.NoTrace("mismatched user address", errors.K.Invalid, "request", request, "tx", tx)
		return
	}

	resp, err = fs.GetRedeemedOffer(tx.ContractAddr, tx.RedeemableId, tx.TokenId)
	if err != nil {
		return
	}
	if resp.Claimed {
		err = errors.NoTrace("token already claimed", errors.K.Invalid, "request", request, "tx", tx)
		return
	}

	var stmt string
	templateArgs := fs.context()
	if stmt, err = mergeTemplate("sql/update-mapping.tmpl", templateArgs); err != nil {
		return
	}

	var args []interface{}
	args = append(args, tx.TokenId)
	args = append(args, tx.ContractAddr)
	args = append(args, tx.RedeemableId)

	var rows *pgx.Rows
	if rows, err = fs.conn().Query(stmt, args...); err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var url, code, addr sql.NullString
		var claimed sql.NullBool
		var created, updated sql.NullTime
		if err = rows.Scan(&url, &code, &addr, &claimed, &created, &updated); err != nil {
			return
		}
		if claimed.Valid && claimed.Bool {
			resp = FulfillmentData{
				ContractAddr: tx.ContractAddr,
				RedeemableId: tx.RedeemableId,
				TokenId:      tx.TokenId,
				Claimed:      true,
				UserAddr:     tx.UserAddr,
				Created:      created.Time,
				Updated:      updated.Time,
				Url:          url.String,
				Code:         code.String,
			}
		}
	}

	return
}

func (fs *FulfillmentPersistence) GetRedeemedOffer(contractAddr, redeemableId, tokenId string) (resp FulfillmentData, err error) {
	var stmt string
	templateArgs := fs.context()
	if stmt, err = mergeTemplate("sql/get-mapping.tmpl", templateArgs); err != nil {
		return
	}

	var args []interface{}
	args = append(args, contractAddr)
	args = append(args, redeemableId)
	args = append(args, tokenId)

	var rows *pgx.Rows
	if rows, err = fs.conn().Query(stmt, args...); err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var id, url, code, addr sql.NullString
		var claimed sql.NullBool
		var created, updated sql.NullTime
		if err = rows.Scan(&id, &url, &code, &addr, &claimed, &created, &updated); err != nil {
			return
		}
		if claimed.Valid && claimed.Bool {
			resp = FulfillmentData{
				ContractAddr: contractAddr,
				RedeemableId: redeemableId,
				TokenId:      tokenId,
				Claimed:      true,
				UserAddr:     addr.String,
				Created:      created.Time,
				Updated:      updated.Time,
				Url:          url.String,
				Code:         code.String,
			}
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
