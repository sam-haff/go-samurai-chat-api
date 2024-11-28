#!/bin/sh

# Build the needed images and prepare all other necessary
# utility files(needed to faciliate deployment update on the server side)
# and upload them to the server using scp command.
#
# Requires:
# - SERVER_BUILD_NUMBER environment variable
# Incremental build version. Should be incremented for the new build.
# - FIREBASE_CREDS_FILE environment variable. 
# A path to firebase service account json. 
# First you need a Firebase project with enabled Authentication.
# Now get the file from Firebase console->Project settings->Service accounts->Generate new private key
# - MONGODB_CONNECT_URL environment variable
# An mongodb connect url in with the credentails.
# mongodb://<username>:<password>@<mongodb_url>/?<options>
# If you use Atlas, you can find it by clicking Connect under your cluster,
# choose Drivers then in any driver step 3 is the connection string which is
# exactly the sought value.
# - SERVER_ADDR environment variable, which
# should contain ip address of the server.
# - Server user <uploaduser> created as described in server_setup_readme.txt
# - SSH key <ssh-key.pem> for the <uploaduser>, you get it from your
# server's /uploadroot/builds_upload/.ssh/id_rsa
# - SSl cert files(fullchain.pem, privkey.pem) 
# for your domain in <proj_root>/nginx/certs directory. You can get them from
# LetsEncrypt ACME or from any other SSL certificates provider.


./build.sh

cp load_images.sh ../images
cp redeploy.sh ../images

./transfer.sh

