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

Here's a sample redemption transaction on our demov3 blockchain:
https://eluvio.tryethernal.com/transaction/0x97686b790728669a54898d812ed59afebab522da03b8161228f0f74cd6693187


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
- response on success: 200
```json
{
  "message": "loaded fulfillment data for a redeemable offer",
  "contract_addr": "0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9",
  "offer_id": "0",
  "url": "https://live.eluv.io/",
  "codes": [
    "ABC123",
    "XYZ789"
  ]
}
```

### Wallet API

- GET `fulfill/:transaction_id`
  - bearer auth token -> user address
  - append `?network=demov3` to lookup transactions on the `demov3` network instead of `main`; GET `fulfill/:transaction_id?network=demov3`
- response on success: 200
```json
{
  "message": "fulfilled redeemable offer",
  "fulfillment_data": {
    "url": "https://live.eluv.io/",
    "code": "XYZ789"
  },
  "transaction": {
    "contract_address": "0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9",
    "user_address": "0xb516b92fe8f422555f0d04ef139c6a68fe57af08",
    "token_id": 34,
    "offer_id": 0
  }
}
```
- response on error or invalid request (eg, tx tokenId already claimed): 400


### Request -> Response Processing

The process is as follows:
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
and it accepts arbitrary data in response.  the daemon then delivers the
metadata + marshalled interface as json.

This splitting of the library and custom code has not been completed.

