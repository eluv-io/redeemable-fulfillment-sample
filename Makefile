#
# Sample Redeemable Offer Fulfillment Service
#

build:
	go build -o bin/fulfillmentd cmd/main.go
	ls -l ./bin/fulfillmentd
	./bin/fulfillmentd

import:
	go mod tidy -compat=1.18

unittest:
	go test -v ./...

logs:
	tail -F ~/ops/logs/fulfillmentd.log

date = $(shell date)
msg = '{ "url": "https://live.eluv.io/", "codes":  [ "ABC123", "XYZ789" ] }'
contract = '0xContract'
redeemable = '0'
tx = 'tx-test-0000'


## select a url
# local dev url
url = http://localhost:2023
# deployed sample url
#url = https://appsvc.svc.eluv.io/res

run:
	@echo "Note: default config requires tunnel to DB on 127.0.0.1:26257"
	./bin/fulfillmentd --config config/config.toml

version:
	curl -s $(url)/version | jq .


#
# these targets require env vars tha contains CF tokens:
#  tok = acspjc...
#  tok_unauth = acspjc...
#

load_codes:
	@echo "todo:"
	curl -s -X POST -d $(msg) -H 'Authorization: Bearer $(tok)' $(url)/load/$(contract)/$(redeemable) | jq .

fulfill_code:
	curl -s -H 'Authorization: Bearer $(tok)' "$(url)/fulfill/$(tx)" | jq .

test_unauthorized:
	@echo "test unauthorized:"
	curl -s -X POST -d $(msg) -H 'Authorization: Bearer $(tok_unauth)' $(url)/fulfill/$(tx) | jq .

test_invalid_user:
	@echo "test invalid user:"
	curl -s -X POST -d $(msg) -H 'Authorization: Bearer $(tok2)' $(url)/fulfill/$(tx) | jq . || true  # direct path only; will be a 404 via nginx


#
# helpers
#

.PHONY: config
config:
	vi  config/config.toml

