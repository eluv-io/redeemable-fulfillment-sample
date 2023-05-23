--
-- DB tables for FulfillmentService
--

CREATE DATABASE IF NOT EXISTS fulfillmentservice;
USE fulfillmentservice;


--- Storage for the library-provided Redeemable Offer Fulfillment Daemon
CREATE TABLE IF NOT EXISTS redeemable_offer_claims (
    id                UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    contract_addr     text NOT NULL,
    redeemable_id     text NOT NULL,
    token_id          text NOT NULL,
    user_addr         text NOT NULL,
    created           timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS roc_id_idx ON redeemable_offer_claims (id);
CREATE INDEX IF NOT EXISTS roc_contract_addr_idx ON redeemable_offer_claims (contract_addr);
CREATE INDEX IF NOT EXISTS roc_redeemable_id_idx ON redeemable_offer_claims (redeemable_id);
CREATE INDEX IF NOT EXISTS roc_token_id_idx ON redeemable_offer_claims (token_id);
CREATE INDEX IF NOT EXISTS roc_claimer_user_addr_idx ON redeemable_offer_claims (user_addr);
CREATE INDEX IF NOT EXISTS roc_created_idx ON redeemable_offer_claims (created);


--- Sample storage for customer's FulfillmentService
CREATE TABLE IF NOT EXISTS fulfillment_service (
    id                UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    --  these are loaded on setup
    contract_addr     text NOT NULL,
    redeemable_id     text NOT NULL,
    url               text NOT NULL,
    code              text NOT NULL,
    --  these are updated by a claim
    claimed           bool NOT NULL DEFAULT false,
    claimer_token_id  text,
    claimer_user_addr text,
    created           timestamptz NOT NULL DEFAULT now(),
    updated           timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS fs_id_idx ON fulfillment_service (id);
CREATE INDEX IF NOT EXISTS fs_contract_addr_idx ON fulfillment_service (contract_addr);
CREATE INDEX IF NOT EXISTS fs_redeemable_id_idx ON fulfillment_service (redeemable_id);
CREATE INDEX IF NOT EXISTS fs_url_idx ON fulfillment_service (url);
CREATE INDEX IF NOT EXISTS fs_code_idx ON fulfillment_service (code);
CREATE INDEX IF NOT EXISTS fs_claimed_idx ON fulfillment_service (claimed);
CREATE INDEX IF NOT EXISTS fs_claimer_token_id_idx ON fulfillment_service (claimer_token_id);
CREATE INDEX IF NOT EXISTS fs_claimer_user_addr_idx ON fulfillment_service (claimer_user_addr);
CREATE INDEX IF NOT EXISTS fs_created_idx ON fulfillment_service (created);
CREATE INDEX IF NOT EXISTS fs_updated_idx ON fulfillment_service (updated);

