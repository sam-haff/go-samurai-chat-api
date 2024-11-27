#!/bin/sh

./build.sh

cp load_images.sh ../images
cp redeploy.sh ../images

./transfer.sh

