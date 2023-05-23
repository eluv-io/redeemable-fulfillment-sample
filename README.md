# redeemable-fulfillment-sample

This repo contains sample code for a service that implements the fulfillment of a NFT redeemable-offer


## Requirements

  - fulfill request - signed message including the redemption transaction ID
     - server verifies that the signer is who redeemed
     - verify not yet fulfilled
     - server calls the fulfillment custom function (whatever needs to be done - our sample can simply write out a file or a db entry)
     - if fulfillment succeeds - write down local state saying this offer is fulfilled


## Sample fulfillment function

This version returns a simple URL + code as the item that is fulfilled.


## Redeemable Offer Docs

https://elv-test-hub.web.app/advanced/tokens/#nft-redeemables


## API

### Setup API

- POST "load/:contract_addr/:redeemable_id"
  - body: `{ "url": URL, "codes": [ list of codes... ] }`
- insert codes into DB as unclaimed


### Wallet API

- GET "fulfill/:transaction_id"
  - bearer auth token -> user address

- response on success: url + code
- response if tx's tokenId has already claimed: 400


### Request -> Response Processing

process as follows:
- extract user address from request
- look up tx on explorer 
  - extract wallet addr, contract addr, tokenId, redeemeableId(bitmask entry)
- verify tx wallet address matches user address
- query DB, verify this contract + redeemableId + tokenId not been redeemed before
- query DB, find matching contract + redeemableId + not-claimed that matches 
   - error if we're out of codes
- mark this tokenId as redeemed in the DB
- return URL and code  (any code can be used for any tokenId)



## Internals: Splitting Eluvio vs Customer service interface

Fulfillment Daemon calls into the customer's fulfillment service interface:
```
FulfillRedeemableOffer(contractAddr, offerId, tokenId, redeemingUserAddress string) string
```
and they return json.  we deliver the metadata + json. (sample returns `{ url, code }`)

This portion has not been completed.
