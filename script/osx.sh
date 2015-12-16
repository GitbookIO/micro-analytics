#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

# Compute repo's dir
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )

# Change to current dir
cd ${DIR}

go build -ldflags "-s" -o "${DIR}/script/build/micro-analytics_darwin_amd64"
