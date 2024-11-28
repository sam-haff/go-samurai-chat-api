#!/bin/bash

# Read server_setup_readme.txt.
# Part of the user setup on the server. 

mkdir /uploadroot
chmod 770 /uploadroot
useradd uploaduser
mkdir /uploadroot/builds_upload
usermod -d /uploadroot/builds_upload
chown uploaduser:uploaduser /uploadroot/builds_upload 
chmod 755 /uploaduser/builds_upload
cd /uploaduser/builds_upload
mkdir new
chown uploaduser:uploaduser new
chmod 755 new
mkdir .ssh
ssh-keygen -t rsa -b 4096 id_rsa
mv id_rsa.pub ./.ssh/authorized_keys
mv id_rsa ./.ssh/id_rsa
chown uploaduser:uploaduser ./.ssh/*
chmod 755 ./.shh/*