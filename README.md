# redeemable-fulfillment-sample

This repo contains sample code for a service that implements the fulfillment of an NFT redeemable-offer.

Redeemable Offers are documented here: https://elv-test-hub.web.app/advanced/tokens/#nft-redeemables

The flow is as follows:
 - server accepts signed message including the redemption transaction ID
 - server verifies that the signer is who redeemed
 - server verifies not yet fulfilled
 - server calls the fulfillment custom function
 - if fulfillment succeeds, persists DB state that this offer is fulfilled

The repo also implements a sample fulfillment -- a simple URL + code, mirroring the case of a coupon (the code) for a website purchase (the URL).


## Getting Started

- clone this repo
- create a `config/config.toml` based on `config/config-example.toml`
- build and run:
```
make build run
```
- smoke test:
```
make version
```


## API

### Setup API

- POST `load/:contract_addr/:redeemable_id`
  - body: `{ "url": URL, "codes": [ list of codes... ] }`
- inserts the codes into DB as unclaimed


### Wallet API

- GET `fulfill/:transaction_id`
  - bearer auth token -> user address

- response on success: url + code
- response on error or invalid request (eg, tx tokenId already claimed): 400


### Request -> Response Processing

process as follows:
- extract user address from request
- look up tx on explorer 
  - extract wallet addr, contract addr, tokenId, redeemeableId(bitmask entry)
- verify tx wallet address matches user address
- query DB, verify this contract + redeemableId + tokenId not been redeemed before
- query DB, find matching contract + redeemableId + not-claimed that matches 
   - error if we're out of codes
- insert this tokenId as redeemed in the DB
- return URL and code (any code can be used for any tokenId)


## Internals: splitting library function vs Customer service interface

The Fulfillment Daemon calls into the customer's fulfillment service 
interface ("fulfillment custom function") with
```
FulfillRedeemableOffer(contractAddr, offerId, tokenId, redeemingUserAddress string) interface{}
```
and it accepts arbitrary data in repsonse.  The daemon then deliver the 
metadata + marshalled interface as json. The sample returns `{ url, code }`.  

This portion has not been completed.

