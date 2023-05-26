//
// Redeemable Offer Fulfillment API interface:
//
// $ curl -s http://localhost:2023/fulfill/:tx
// {
//  "message": "fulfilled redeemable offer",
//  "fulfillment_data": {
//    "url": "https://eluv.io/",
//    "code": "XYZ789"
//  },
//  "transaction": {
//    "contract_address": "0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9",
//    "user_address": "0xb516b92fe8f422555f0d04ef139c6a68fe57af08",
//    "token_id": 34,
//    "offer_id": 0
//  }
//}
//
// $ curl -s http://localhost:2023/load/:token_addr/:redeemable_id --data '{ "url": "https://eluv.io/", "codes": [ "ABC123", "XYZ789" ] }'
// {
//  "message": "loaded fulfillment data for a redeemable offer",
//  "contract_addr": "0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9",
//  "offer_id": "0",
//  "url": "https://eluv.io/",
//  "codes": [ "ABC123", "XYZ789" ]
//}

package api

import (
	"fulfillmentd/redeemservice/db"
	"fulfillmentd/server"
	"fulfillmentd/utils"
	elog "github.com/eluv-io/log-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

var log = elog.Get("/fs/api")

type FulfillmentResponse struct {
	Message         string                   `json:"message"`
	FulfillmentData interface{}              `json:"fulfillment_data"`
	Transaction     db.RedemptionTransaction `json:"transaction"`
}

type LoadRequest struct {
	Url   string   `json:"url"`
	Codes []string `json:"codes"`
}

type LoadResponse struct {
	Message      string   `json:"message"`
	ContractAddr string   `json:"contract_addr"`
	OfferId      string   `json:"offer_id"`
	Url          string   `json:"url"`
	Codes        []string `json:"codes"`
}

func AddRoutes(s *server.Server) {
	log.Info("Adding FS routes")
	public := s.Router.Group("/")
	public.POST("load/:contract_addr/:redeemable_id", LoadFulfillmentData(s.FulfillmentService))
	public.GET("fulfill/:transaction_id", FulfillRedeemableOffer(s.FulfillmentService))
}

func LoadFulfillmentData(fs *server.FulfillmentService) gin.HandlerFunc {
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
			ContractAddress: contractAddr,
			OfferId:         redeemableId,
			Url:             data.Url,
			Codes:           data.Codes,
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
			OfferId:      redeemableId,
			Url:          data.Url,
			Codes:        data.Codes,
		}
		ctx.JSON(http.StatusOK, ret)
	}
}

// FulfillRedeemableOffer godoc
// @ID offer-redemption
// @Summary FulfillRedeemableOffer
// @Description FulfillRedeemableOffer
// @Param transaction_id path string true "blockchain transaction id that shows the redeemable offer was redeemed"
// @Param network query string false "which ELV network to look up transaction: 'main' or 'demov3'; defaults to main"
// @Produce  json
// @Router /fulfill/:transaction_id [GET]
func FulfillRedeemableOffer(fs *server.FulfillmentService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error

		var request db.FulfillmentRequest
		if err = ctx.ShouldBind(&request); err != nil {
			log.Warn("error binding request body", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "error binding request body"})
		}

		request.Transaction = ctx.Param("transaction_id")
		request.UserAddress, err = utils.ExtractUserAddress(ctx)
		if err != nil {
			log.Warn("error extracting user address", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error extracting user address",
				"err":     err,
			})
			return
		}

		request.Network = utils.IfElse(ctx.Query("network") == "", "main", ctx.Query("network"))

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
			FulfillmentData: struct {
				Url  string `json:"url"`
				Code string `json:"code"`
			}{
				Url:  data.Url,
				Code: data.Code,
			},
			Transaction: data.ToTransaction(),
		}
		ctx.JSON(http.StatusOK, ret)
	}
}
