#
# Sample Redeemable Offer Fulfillment Service
#

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

date=$(shell date)
msg='{ "url": "https://live.eluv.io/", "codes":  [ "ABC123", "XYZ789" ] }'
# Goat One on demov3
contract=0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9
offerId=0
test_tx=tx-test-0000
tx=0x97686b790728669a54898d812ed59afebab522da03b8161228f0f74cd6693187
h= -H "Content-Type: application/json"

## select a url
# local dev url
url = http://localhost:2023
# deployed sample url
#url = https://appsvc.svc.eluv.io/codefulfillment

run:
	@echo "Note: default config requires tunnel to DB on 127.0.0.1:26257"
	./bin/fulfillmentd --config config/config.toml

build_and_run_with_logs:
	( make build && (make run & sleep 2 && make logs))

#
# these targets require an env vars tha contains a CF token:
#  tok = acspjc...
#

load_codes:
	curl -s -X POST $h -d $(msg) -H 'Authorization: Bearer $(tok)' $(url)/load/0x0/$(offerId) | jq .
	curl -s -X POST $h -d $(msg) -H 'Authorization: Bearer $(tok)' $(url)/load/$(contract)/$(offerId) | jq .

test_fulfill_code:
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/$(test_tx)?network=demov3" | jq .

fulfill_code:
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/$(tx)?network=demov3" | jq .

test_invalid_user:
	@echo "test invalid user:"
	curl -s -H 'Authorization: Bearer $(tok)' $(url)/fulfill/tx-test-invaliduser?network=demov3 | jq .

test_out_of_codes:
	@echo "use after load_codes"
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/tx-test-0000?network=demov3" | jq .
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/tx-test-0001?network=demov3" | jq .
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/tx-test-0002?network=demov3" | jq .

#
# helpers
#

.PHONY: config version
config:
	vi  config/config.toml

version:
	curl -s $(url)/version | jq .

