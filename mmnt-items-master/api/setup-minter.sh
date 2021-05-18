#!/usr/bin/env bash

# Setup minter account

curl --request POST \
  --url http://localhost:3000/v1/moments/setup \
  --header 'Content-Type: application/json'

curl --request POST \
  --url http://localhost:3000/v1/mmnt-items/setup \
  --header 'Content-Type: application/json'

curl --request POST \
  --url http://localhost:3000/v1/market/setup \
  --header 'Content-Type: application/json'

# Mint Moment and Mmnt Items

curl --request POST \
  --url http://localhost:3000/v1/moments/mint \
  --header 'Content-Type: application/json' \
  --data '{
    "recipient": "'${FLOW_ADDRESS}'",
    "amount": 50.0
  }'

curl --request POST \
  --url http://localhost:3000/v1/mmnt-items/mint \
  --header 'Content-Type: application/json' \
  --data '{
    "recipient": "'${FLOW_ADDRESS}'",
    "typeID": 1
  }'

curl --request POST \
  --url http://localhost:3000/v1/market/sell \
  --header 'Content-Type: application/json' \
  --data '{
    "itemID": 0,
    "price": 7.5
  }'
