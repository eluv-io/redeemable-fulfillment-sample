package utils

import (
	"fmt"
	"github.com/eluv-io/common-go/format/eat"
	"github.com/eluv-io/common-go/util/ethutil"
	"github.com/eluv-io/errors-go"
	"github.com/eluv-io/log-go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"net/http"
	"runtime"
	"strings"
)

func IfElse[T any](cond bool, trueVal, falseVal T) T {
	if cond {
		return trueVal
	}
	return falseVal
}

func ExtractUserAddress(ctx *gin.Context) (addr string, err error) {
	e := errors.TemplateNoTrace("ExtractUserAddress", errors.K.Invalid)

	authHeader := ctx.Request.Header["Authorization"]
	if authHeader == nil {
		err = e("invalid Auth: missing header", authHeader)
		return
	}
	split := strings.Split(authHeader[0], " ")
	if len(split) < 2 {
		err = e("invalid Auth: invalid header", authHeader)
		return
	}

	var tok *eat.Token
	tok, err = ParseAuthToken(ctx.Request)
	if err != nil {
		return
	}
	addr = strings.ToLower(tok.EthAddr.Hex())
	//log.Trace("ExtractUserAddress", "addr", addr, "json", tok.AsJSON())

	return
}

func CompareAuth(ctx *gin.Context) (isValid bool, err error) {
	e := errors.TemplateNoTrace("check auth", errors.K.Invalid)

	addr := ctx.Param("addr")
	authHeader := ctx.Request.Header["Authorization"]
	if authHeader == nil {
		return false, e("invalid Auth: missing header", authHeader)
	}
	split := strings.Split(authHeader[0], " ")
	if len(split) < 2 {
		return false, e("invalid Auth: invalid header", authHeader)
	}

	tok, err := ParseAuthToken(ctx.Request)
	if err != nil {
		//log.Error("error in ParseAuthToken", err)
		return false, err
	}
	log.Debug("Token:", "json", tok.AsJSON())

	if AddressMatches(tok, addr) {
		return true, nil
	}
	return false, e("address mismatch", addr+" != "+tok.EthAddr.Hex())
}

// NormalizeAddress : Normalize address string `addr` in either hex or subject format
// return : lowercase hex string, or empty string if it cannot be parsed
func NormalizeAddress(addr string) string {
	var a common.Address
	var e error
	if a, e = ethutil.HexToAddress(addr); e != nil {
		if a, e = ethutil.IDStringToAddress(addr); e != nil {
			return ""
		}
	}
	log.Trace("normalized address:", "hex", a.Hex())
	return strings.ToLower(a.Hex())
}

// AddressMatches : verify address `addr` matches either the Hex address or Subject ID in `token`
func AddressMatches(token *eat.Token, addr string) bool {
	var a common.Address
	var e error
	if a, e = ethutil.HexToAddress(addr); e != nil {
		if a, e = ethutil.IDStringToAddress(addr); e != nil {
			return false
		}
	}
	log.Trace("compare", "tok.hex", token.EthAddr.Hex(), "tok.subj", token.SID.String(), "?= addr", addr)
	if a == token.EthAddr || addr == token.Subject {
		return true
	}
	return false
}

// ParseAuthToken pulls the bearer auth token out of `r` and parses it
// Code borrowed from elv-master, could move to common-go and share.
func ParseAuthToken(r *http.Request) (*eat.Token, error) {
	// Bearer Token Authorization: https://tools.ietf.org/html/rfc6750
	var token64 string

	// Authorization Request Header Field: Authorization: Bearer mF_9.B5f-4.1JqM
	hdr := r.Header.Get("Authorization")
	if len(hdr) > 0 {
		split := strings.Split(hdr, " ")
		if len(split) != 2 {
			return nil, errors.Str("malformed authorization header")
		}
		token64 = strings.TrimSpace(split[1])
	} else {
		// URI Query Parameter: GET /resource?access_token=mF_9.B5f-4.1JqM HTTP/1.1
		//
		// "Because of the security weaknesses associated with the URI method ... it
		// SHOULD NOT be used unless it is impossible to transport the access token
		// in the "Authorization" Request header field or the HTTP Request
		// entity-body."
		//
		// Note: OAuth uses "access_token"
		token64 = r.URL.Query().Get("authorization")
	}

	if len(token64) == 0 {
		return nil, errors.Str("missing authorization token")
	}

	return ParseClientBearerToken(token64)
}

// ParseClientBearerToken parses a given bearer auth token `authToken`
func ParseClientBearerToken(authToken string) (*eat.Token, error) {
	return eat.Parse(authToken)
}

func ReturnError(ctx *gin.Context, httpError int, err error) {
	if err == nil {
		err = errors.E("aborted")
	}

	var pgErr *pgx.PgError
	if errors.As(err, &pgErr) {
		err = errors.E(err, "db-error-code", pgErr.Code, "db-error-msg", pgErr.Message)
	}

	err = errors.ClearStacktrace(err)
	log.Debug("http-error", "code", httpError, "where", where(1), "err", err)
	body := gin.H{
		"error": gin.H{
			"code":    httpError,
			"message": err.Error(),
		},
	}
	ctx.JSON(httpError, body)
}

// caller logging helpers

func where(extraSkip int) string {
	file, line, name := trace(extraSkip + 2)
	return fmt.Sprintf("%s:%d:%s", file, line, name)
}

func trace(skip int) (string, int, string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "?", 0, "?"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return file, line, "?"
	}

	names := strings.Split(fn.Name(), ".")
	name := names[len(names)-1]

	files := strings.Split(file, "/")
	file = files[len(files)-1]

	return file, line, name + "()"
}
