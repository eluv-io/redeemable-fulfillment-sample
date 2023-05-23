//
// Redeemable Offer Fulfillment API interface:
//
// $ curl -s http://localhost:2023/fulfill/:tx
//  {
//    "message": "fulfilled redeemable offer",
//    "url": "https://live.eluv.io/",
//    "code": "UPv7uzPs",
//  }
//
// $ curl -s http://localhost:2023/load/:token_addr/:redeemable_id --data '{ "url": "https://live.eluv.io/", "codes": [ "ABC123", "XYZ789" ] }'
//  {
//    "message": "loaded fulfillment data for a redeemable offer",
//    "token_addr": "0x....",
//    "redeemable_id": 0,
//  }

package api

import (
	"fulfillmentd/authority"
	"fulfillmentd/redeemservice/db"
	"fulfillmentd/utils"
	elog "github.com/eluv-io/log-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

var log = elog.Get("/fs/api")

type FulfillmentResponse struct {
	Message string `json:"message"`
	Url     string `json:"url"`
	Code    string `json:"code"`
}

type LoadRequest struct {
	Url   string   `json:"url"`
	Codes []string `json:"codes"`
}

type LoadResponse struct {
	Message      string `json:"message"`
	ContractAddr string `json:"contract_addr"`
	RedeemableId string `json:"redeemable_id"`
}

func AddRoutes(s *authority.Server) {
	log.Info("Adding FS routes")
	public := s.Router.Group("/")
	public.POST("load/:contract_addr/:redeemable_id", LoadFulfillmentData(s.FulfillmentService))
	public.GET("fulfill/:transaction_id", FulfillRedeemableOffer(s.FulfillmentService))
}

func LoadFulfillmentData(fs *authority.FulfillmentService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error

		var data LoadRequest
		if err = ctx.ShouldBind(&data); err != nil {
			log.Warn("error binding request body", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "error binding request body"})
		}

		contractAddr := ctx.Param("contract_addr")
		redeemableId := ctx.Param("redeemable_id")

		setupData := db.SetupData{
			ContractAddr: contractAddr,
			RedeemableId: redeemableId,
			Url:          data.Url,
			Codes:        data.Codes,
		}
		if err = fs.SetupFulfillment(setupData); err != nil {
			log.Debug("error with data setup", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error loading fulfillment data",
				"err":     err,
			})
			return
		}

		ret := LoadResponse{
			Message:      "loaded fulfillment data for a redeemable offer",
			ContractAddr: contractAddr,
			RedeemableId: redeemableId,
		}
		ctx.JSON(http.StatusOK, ret)
	}
}

func FulfillRedeemableOffer(fs *authority.FulfillmentService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error

		var request db.FulfillmentRequest
		if err = ctx.ShouldBind(&request); err != nil {
			log.Warn("error binding request body", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "error binding request body"})
		}

		request.Transaction = ctx.Param("transaction_id")
		request.UserAddr, err = utils.ExtractUserAddress(ctx)
		if err != nil {
			log.Warn("error extracting user address", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error extracting user address",
				"err":     err,
			})
			return
		}

		var data db.FulfillmentData
		data, err = fs.FulfillRedeemableOffer(request)
		if err != nil {
			log.Debug("error fulfilling offer", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error fulfilling offer",
				"err":     err,
			})
			return
		}

		ret := FulfillmentResponse{
			Message: "fulfilled redeemable offer",
			Url:     data.Url,
			Code:    data.Code,
			//ContractAddr: "", RedeemableId: "", TokenId:"",
		}
		ctx.JSON(http.StatusOK, ret)
	}
}
