#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

# Compute repo's dir
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )

# Change to current dir
cd ${DIR}

# Copy Dockerfile
cp script/linux64.Dockerfile ./Dockerfile

##
# Build
##

# Build tmp image
image_id=$(docker build . | tail -n 1 | cut -f3 -d' ')
echo "Image: ${image_id}"

# Create tmp container
container_id=$(docker create ${image_id})
echo "Container: ${container_id}"

# Copy out prebuilt source
docker cp "${container_id}:/micro-analytics_linux_amd64" "${DIR}/script/build/"

##
# Cleanup
##

# Remove tmp docker container
docker rm -f "${container_id}"

# Remove tmp docker image
docker rmi -f "${image_id}"

# Remove copied Dockerfile
rm "${DIR}/Dockerfile"
