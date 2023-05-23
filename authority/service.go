package authority

import (
	"fulfillmentd/redeemservice/db"
)

type FulfillmentService struct {
	db *db.FulfillmentPersistence
}

func NewFulfillmentService(s *Server) *FulfillmentService {
	return &FulfillmentService{
		db: db.NewFulfillmentPersistence(s.ConnectionManager),
	}
}

func (fs *FulfillmentService) SetupFulfillment(data db.SetupData) (err error) {
	return fs.db.SetupFulfillment(data)
}

func (fs *FulfillmentService) FulfillRedeemableOffer(request db.FulfillmentRequest) (data db.FulfillmentData, err error) {
	return fs.db.FulfillRedeemableOffer(request)
}
