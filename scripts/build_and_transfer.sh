#!/bin/sh

# Build the needed images and prepare all other necessary
# utility files(needed to faciliate deployment update on the server side)
# and upload them to the server using scp command.

./build.sh

cp load_images.sh ../images
cp redeploy.sh ../images

./transfer.sh

