#!/bin/bash

if [[ -z "$FIREBASE_CREDS_FILE" || -z "$MONGODB_CONNECT_URL" ]]; then
  echo "Error: env vars are not set"
  exit 1
fi

touch ../.env
echo "FIREBASE_CREDS_FILE=$FIREBASE_CREDS_FILE" >> ../.env
echo "MONGODB_CONNECT_URL=$MONGODB_CONNECT_URL" >> ../.env