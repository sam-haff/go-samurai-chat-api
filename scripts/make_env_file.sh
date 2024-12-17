#!/bin/bash

# Make .env file from CI/CD secrets.
# Requires:
# - FIREBASE_CREDS_FILE env variable. 
# A path to firebase service account json. 
# First you need a Firebase project with enabled Authentication.
# Now get the file from Firebase console->Project settings->Service accounts->Generate new private key
# - MONGODB_CONNECT_URL env variable
# An mongodb connect url in with the credentails.
# mongodb://<username>:<password>@<mongodb_url>/?<options>
# If you use Atlas, you can find it by clicking Connect under your cluster,
# choose Drivers then in any driver step 3 is the connection string which is
# exactly the sought value.

if [[ -z "$FIREBASE_CREDS_FILE" || -z "$MONGODB_CONNECT_URL" ]]; then
  echo "Error: env vars are not set(make_env_file)"
  exit 1
fi

echo "FIREBASE_CREDS_FILE=$FIREBASE_CREDS_FILE" >> ../.env
echo "MONGODB_CONNECT_URL=$MONGODB_CONNECT_URL" >> ../.env
echo "FIREBASE_STORAGE_BUCKET=$FIREBASE_STORAGE_BUCKET" >> ../.env