package db

import (
	"context"
	"fmt"
	"github.com/eluv-io/contracts/contracts-go/tradable"
	"github.com/eluv-io/errors-go"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"strings"
)

// ToRedeemable converts based on https://gist.github.com/elv-preethi/44e0a809d3e7daa4e7713d6b23ead136
func (fp *FulfillmentPersistence) ToRedeemable(fr FulfillmentRequest) (redemption RedemptionTransaction, err error) {
	// get data from tx
	log.Debug("using eth network", "network", fr.Network)
	var ec *ethclient.Client
	ec, err = ethclient.Dial(fp.ethUrlByNetwork[fr.Network])
	if err != nil {
		return
	}
	defer ec.Close()

	var receipt *types.Receipt
	receipt, err = ec.TransactionReceipt(context.Background(), common.HexToHash(fr.Transaction))
	if err != nil {
		return
	}

	var tr *tradable.ElvTradableRedeem
	if len(receipt.Logs) > 0 {
		var instance *tradable.ElvTradable
		instance, err = tradable.NewElvTradable(receipt.Logs[0].Address, ec)
		if err != nil {
			return
		}

		tr, err = instance.ParseRedeem(*receipt.Logs[0])
		if err != nil {
			return
		}
	} else {
		err = errors.NoTrace("no logs found in receipt", errors.K.Invalid, "receipt", receipt)
		return
	}

	contractAddress := receipt.Logs[0].Address.String()
	hash := common.BytesToHash(common.FromHex(fr.Transaction))
	var isPending bool
	_, isPending, err = ec.TransactionByHash(context.Background(), hash)
	if err != nil {
		err = errors.NoTrace("cannot find tx", "err", err)
		return
	}
	if isPending {
		err = errors.NoTrace("tx is pending", errors.K.Invalid)
		return
	}

	redemption = RedemptionTransaction{
		ContractAddress: strings.ToLower(contractAddress),
		RedeemerAddress: strings.ToLower(tr.Redeemer.String()),
		TokenId:         tr.TokenId.Int64(),
		OfferId:         tr.OfferId,
		IsPending:       isPending,
	}
	log.Info("ToRedeemable", "redemption", fmt.Sprintf("%+v", redemption))

	return
}

// resolveTransactionData does an external query to the ELV blockchain to resolve the data from in the request transaction.
// It also provides mock data for testing from `make load_codes` + `make fulfill_code`
func (fp *FulfillmentPersistence) resolveTransactionData(request FulfillmentRequest) (rt RedemptionTransaction, err error) {
	var isTestData bool
	isTestData, rt = fp.fillTestData(request)
	if isTestData {
		return
	}

	log.Info("resolveTransactionData", "isTestData", isTestData, "redemption", fmt.Sprintf("%+v", rt))
	rt, err = fp.ToRedeemable(request)
	if err != nil {
		return
	}

	log.Info("resolveTransactionData", "isTestData", isTestData, "redemption", fmt.Sprintf("%+v", rt))
	return
}

// fillTestData is for integration testing. It provides mock data for testing from `make load_codes` + `make fulfill_code`
func (fp *FulfillmentPersistence) fillTestData(request FulfillmentRequest) (isTestData bool, data RedemptionTransaction) {
	isTestData = false
	if strings.Contains(request.Transaction, "tx-test") {
		isTestData = true
		testTx := request.Transaction

		request.Transaction = "0x6ba5f67b3c477422260808f3120a6b2efec9453d167661c171a3501e65f9d29d"
		log.Warn("converting tx-test to tx", "tx", request.Transaction)

		var err error
		data, err = fp.ToRedeemable(request)
		if err != nil {
			log.Error("cannot convert tx-test to tx", "err", err)
			return
		}

		switch testTx {
		case "tx-test-0000":
			data.RedeemerAddress = request.UserAddress
			data.TokenId = 1
		case "tx-test-0001":
			data.RedeemerAddress = request.UserAddress
			data.TokenId = 2
		case "tx-test-0002":
			data.RedeemerAddress = request.UserAddress
			data.TokenId = 3
		case "tx-test-invaliduser":
			// already invalid
		}
	}

	if isTestData {
		log.Warn("forged redemption data", "redemption", data)
	}

	return
}
