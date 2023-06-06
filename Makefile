#
# Sample Redeemable Offer Fulfillment Service
#

# Select a url
# local dev:
url = http://localhost:2023
# deployed sample:
url = https://appsvc.svc.eluv.io/main/code-fulfillment


# NFT w/ offer: Goat One on demov3
contract=0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9
offerId=0
# Sample redeem tx -- https://eluvio.tryethernal.com/transaction/0x7f48187a55836aa0a7da0ff591e8d34e5fce1075725e3bba6ec041b6f9d5fc8e
tx=0x7f48187a55836aa0a7da0ff591e8d34e5fce1075725e3bba6ec041b6f9d5fc8e

# NFT w/ offer: Acme Ticket on demov3
contract2=0x2d9729b9f7049bb3cd6c4ed572f7e6f47922ca68

# eluvio prod sheep
#contract3=0xe70d12af413a3a4caf2e8e182560c7324268b443
#contract3=0xd4c8153372b0292b364dac40d0ade37da4c4869a

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
	curl -s -H 'Authorization: Bearer $(tok)' $(url)/demov3/fulfill/tx-invaliduser | jq .

test_invalid_network:
	@echo "test invalid user:"
	curl -s -H 'Authorization: Bearer $(tok)' $(url)/invalid/fulfill/$(tx) | jq .

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

