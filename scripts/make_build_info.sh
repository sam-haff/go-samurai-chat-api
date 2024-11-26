#!/bin/bash

if [[ -z "$SERVER_BUILD_NUMBER" ]] ; then
    echo "Error: Env vars not set"
    exit 1
fi

touch ../images/build_info.json
echo "{ \"version\":"$SERVER_BUILD_NUMBER" }" >> ../images/build_info.json