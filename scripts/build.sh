#!/bin/bash

./build_chat_api.sh
./build_reverse_proxy.sh
./make_build_info.sh
cp ..\docker-compose.yaml ..\images\