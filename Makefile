#
# Sample Redeemable Offer Fulfillment Service
#

# Select a url
# local dev:
url = http://localhost:2023
# deployed sample:
#url = https://appsvc.svc.eluv.io/code-fulfillment


# NFT w/ offer: Goat One on demov3
contract=0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9
offerId=0
# Sample redeem tx -- https://eluvio.tryethernal.com/transaction/0x7f48187a55836aa0a7da0ff591e8d34e5fce1075725e3bba6ec041b6f9d5fc8e
tx=0x7f48187a55836aa0a7da0ff591e8d34e5fce1075725e3bba6ec041b6f9d5fc8e

test_tx=tx-test-0000
msg='{ "url": "https://eluv.io/", "codes":  [ "ABC123", "XYZ789" ] }'
h= -H "Content-Type: application/json"


build:
	(cd version ; go generate)
	go build -o bin/fulfillmentd cmd/main.go
	ls -l ./bin/fulfillmentd
	./bin/fulfillmentd

import:
	go mod tidy -compat=1.18

unittest:
	go test -v ./...

logs:
	tail -F ~/ops/logs/fulfillmentd.log

run:
	@echo "Note: sample config uses tunnel to DB on 127.0.0.1:26257"
	./bin/fulfillmentd --config config/config.toml

build_and_run_with_logs:
	( make build && (make run & sleep 2 && make logs))

#
# these targets require an env vars that contains a CF token:
#  export tok=acspjc...
#

load_codes:
	curl -s -X POST $h -d $(msg) -H 'Authorization: Bearer $(tok)' $(url)/demov3/load/0x0/$(offerId) | jq .
	curl -s -X POST $h -d $(msg) -H 'Authorization: Bearer $(tok)' $(url)/demov3/load/$(contract)/$(offerId) | jq .

test_fulfill_code:
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/demov3/fulfill/$(test_tx)" | jq .

#
# this targets requires an env var that contains the transaction, and the matching user token:
#  export tok=acspjc...
#  export tx=0x7f48187a55836aa0a7da0ff591e8d34e5fce1075725e3bba6ec041b6f9d5fc8e
#
fulfill_code:
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/demov3/fulfill/$(tx)" | jq .

test_invalid_user:
	@echo "test invalid user:"
	curl -s -H 'Authorization: Bearer $(tok)' $(url)/demov3/fulfill/tx-test-invaliduser | jq .

test_out_of_codes:
	@echo "use after load_codes"
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/demov3/fulfill/tx-test-0000" | jq .
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/demov3/fulfill/tx-test-0001" | jq .
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/demov3/fulfill/tx-test-0002" | jq .

#
# helpers
#

.PHONY: config version
config:
	vi  config/config.toml

version:
	curl -s $(url)/version | jq .
	curl -s $(url)/main/version | jq .
	curl -s $(url)/demov3/version | jq .

