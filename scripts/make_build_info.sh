#!/bin/bash

# Make build_info.json file, needed for the server 
# to know whether the redeployemnt is needed.
# Requires $SERVER_BUILD_NUMBER env variable,
# which should contain incremental build version.

if [[ -z "$SERVER_BUILD_NUMBER" ]] ; then
    echo "Error: Env vars not set"
    exit 1
fi

echo "{ \"version\":"$SERVER_BUILD_NUMBER" }" >> ../images/build_info.json