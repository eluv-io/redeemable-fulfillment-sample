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
	"fulfillmentd/constants"
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
	public.POST(":network/load/:contract_addr/:redeemable_id", LoadFulfillmentData(s.FulfillmentService))
	public.GET(":network/fulfill/:transaction_id", FulfillRedeemableOffer(s.FulfillmentService))
}

// LoadFulfillmentData godoc
// @ID offer-redemption-load
// @Summary Load fulfillment data for a redeemable offer
// @Description Load fulfillment data for a redeemable offer
// @Param network path string true "which ELV network the contract is on: 'main' or 'demov3'.  This is ignored for now."
// @Param contract_addr path string true "the contract address of the redeemable offer"
// @Param redeemable_id path string true "the redeemable offer id"
// @Param load_request body LoadRequest true "the fulfillment data url and codes to load"
// @Produce  json
// @Router /:network/load/:contact_addr/:redeemable_id [POST]
func LoadFulfillmentData(fs *server.FulfillmentService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error

		network := ctx.Param("network")
		log.Info("LoadFulfillmentData ignores network for now", "network", network)

		var loadRequest LoadRequest
		if err = ctx.ShouldBind(&loadRequest); err != nil {
			log.Warn("error binding request body", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "error binding request body"})
		}

		contractAddr := ctx.Param("contract_addr")
		redeemableId := ctx.Param("redeemable_id")

		setupData := db.SetupData{
			ContractAddress: contractAddr,
			OfferId:         redeemableId,
			Url:             loadRequest.Url,
			Codes:           loadRequest.Codes,
		}
		if err = fs.SetupFulfillment(setupData); err != nil {
			log.Debug("error with loadRequest setup", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error loading fulfillment loadRequest",
				"err":     err,
			})
			return
		}

		ret := LoadResponse{
			Message:      "loaded fulfillment loadRequest for a redeemable offer",
			ContractAddr: contractAddr,
			OfferId:      redeemableId,
			Url:          loadRequest.Url,
			Codes:        loadRequest.Codes,
		}
		ctx.JSON(http.StatusOK, ret)
	}
}

// FulfillRedeemableOffer godoc
// @ID offer-redemption
// @Summary FulfillRedeemableOffer
// @Description FulfillRedeemableOffer
// @Param network path string true "which ELV network to look up transaction: 'main' or 'demov3'"
// @Param transaction_id path string true "blockchain transaction id that shows the redeemable offer was redeemed"
// @Produce  json
// @Router /:network/fulfill/:transaction_id [GET]
func FulfillRedeemableOffer(fs *server.FulfillmentService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error

		var request db.FulfillmentRequest
		request.Transaction = ctx.Param("transaction_id")
		request.Network = ctx.Param("network")
		switch request.Network {
		case constants.Main, constants.Demov3:
			// ok
		default:
			log.Warn("invalid network", "network", request.Network)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid network",
				"network": request.Network,
				"err":     "invalid elv network name; expected main or demov3",
			})
			return
		}
		request.UserAddress, err = utils.ExtractUserAddress(ctx)
		if err != nil {
			log.Warn("error extracting user address", "err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "error extracting user address",
				"err":     err,
			})
			return
		}

		var fulfillment db.FulfillmentResponse
		fulfillment, err = fs.FulfillRedeemableOffer(request)
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
				Url:  fulfillment.Url,
				Code: fulfillment.Code,
			},
			Transaction: fulfillment.ToTransaction(),
		}
		ctx.JSON(http.StatusOK, ret)
	}
}
