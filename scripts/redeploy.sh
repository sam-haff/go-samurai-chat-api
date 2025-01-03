#!/bin/bash

# Script for redeployment on the server side.
# Assumes the following folder structure:
#-/uploadroot
#   -builds_upload
#       -running
#           *contains files from -new folder
#       -new
#           *folder to which we are uploading files in the <deploy> CI/CD job
# Script checks whether the -running folder contains
# the newest version of the build by comparing <version>
# in /running/build_info.json and /new/build_info.json.
# If version number in /new folder is higher than it
# is in /running then shuts down the currently running
# containers, copies all files from /new to /running
# and restarts the containers.
#
# Can be used as a cron job.

cd /uploadroot/builds_upload/new

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
        docker rmi -f $(docker images -aq)
        ./load_images.sh
        docker compose up -d

        return 0
    fi

    cd ../running
    if [ -z `docker ps -q --no-trunc | grep $(docker compose ps -q 'api')` ]; then 
        echo "Server is not running, starting up..."
        docker compose down # just to be sure
        docker rmi -f $(docker images -aq) #TODO: check if images exist, if so dont reload them
        ./load_images.sh
        docker compose up -d

        return $?
    fi

    return 1
}

deploy