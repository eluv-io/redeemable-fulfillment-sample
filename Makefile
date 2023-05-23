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
contract=0xContractAddress
redeemable=0
tx=tx-test-0000
h= -H "Content-Type: application/json"

## select a url
# local dev url
url = http://localhost:2023
# deployed sample url
#url = https://appsvc.svc.eluv.io/res

run:
	@echo "Note: default config requires tunnel to DB on 127.0.0.1:26257"
	./bin/fulfillmentd --config config/config.toml


#
# these targets require an env vars tha contains a CF token:
#  tok = acspjc...
#

load_codes:
	@echo "todo:"
	curl -s -X POST $h -d $(msg) -H 'Authorization: Bearer $(tok)' $(url)/load/$(contract)/$(redeemable) | jq .

fulfill_code:
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/$(tx)" | jq .

test_invalid_user:
	@echo "test invalid user:"
	curl -s -H 'Authorization: Bearer $(tok)' $(url)/fulfill/tx-test-invaliduser | jq .

test_out_of_codes:
	@echo "use after load_codes"
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/tx-test-0000" | jq .
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/tx-test-0001" | jq .
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/tx-test-0002" | jq .

#
# helpers
#

.PHONY: config version
config:
	vi  config/config.toml

version:
	curl -s $(url)/version | jq .

