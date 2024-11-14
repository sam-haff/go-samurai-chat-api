#!/bin/bash

./build_chat_api.sh
./build_reverse_proxy.sh
cp ..\docker-compose.yaml ..\images\