--
-- DB tables for FulfillmentService
--

CREATE DATABASE IF NOT EXISTS fulfillmentservice;
USE fulfillmentservice;


--- Aggregate table for combined service: library storage plus code+url fulfillment
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
CREATE INDEX IF NOT EXISTS fs_contract_addr_idx ON fulfillment_service (contract_addr);
CREATE INDEX IF NOT EXISTS fs_claimer_user_addr_idx ON fulfillment_service (claimer_user_addr);


--- Storage for a library-provided Redeemable Offer Fulfillment Daemon accepted claims
CREATE TABLE IF NOT EXISTS redeemable_offer_claims (
    id                UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    contract_addr     text NOT NULL,
    offer_id          text NOT NULL,
    token_id          text NOT NULL,
    user_addr         text NOT NULL,
    created           timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS roc_contract_addr_idx ON redeemable_offer_claims (contract_addr);
CREATE INDEX IF NOT EXISTS roc_claimer_user_addr_idx ON redeemable_offer_claims (user_addr);
