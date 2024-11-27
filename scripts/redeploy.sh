#!/bin/bash

function is_new_version() {
    local new_version=`jq '.version' ./build_info.json`
    local running_version=`jq '.version' ../running/build_info.json`

    if [ "$new_version" != "$running_version" ] ; then
        return 1
    fi

    return 0
}

function deploy() {
    if [ ! -d "../running" ]; then
        mkdir ../running
        cp * ../running
    fi

    is_new_version

    if [ $? -ne 0 ]; then
        echo "New version is available, restarting the server..."

        cd ../running
        docker compose down
        cp ../new/* ./
        ./load_images.sh
        docker compose up -d

        return 0
    fi

    cd ../running
    if [ -z `docker ps -q --no-trunc | grep $(docker-compose ps -q 'api')` ]; then 
        echo "Server is not running, starting up..."
        docker compose down # just to be sure
        ./load_images
        docker compose up -d

        return $?
    fi

    return 1
}

deploy