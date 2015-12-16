#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

# Compute script's dir
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Switch to script's dir
cd "${DIR}"

echo "Building OSX"
./osx.sh

echo "Building Linux 64"
./linux64.sh
