package server

import (
	"fulfillmentd/redeemservice/db"
)

type FulfillmentService struct {
	db *db.FulfillmentPersistence
}

func NewFulfillmentService(s *Server) *FulfillmentService {
	return &FulfillmentService{
		db: db.NewFulfillmentPersistence(s.ConnectionManager, s.Cfg.EthUrlsByNetwork),
	}
}

func (fs *FulfillmentService) AvailableNetworks() (nets []string) {
	return fs.db.AvailableNetworks()
}

func (fs *FulfillmentService) SetupFulfillment(setup db.SetupData) (err error) {
	return fs.db.SetupFulfillment(setup)
}

func (fs *FulfillmentService) FulfillRedeemableOffer(request db.FulfillmentRequest) (fd db.FulfillmentResponse, err error) {
	return fs.db.FulfillRedeemableOffer(request)
}

func (fs *FulfillmentService) GetRedeemableOffer(contractAddr, redeemableId, tokenId string) (fd db.FulfillmentResponse, err error) {
	return fs.db.GetRedeemedOffer(contractAddr, redeemableId, tokenId)
}
