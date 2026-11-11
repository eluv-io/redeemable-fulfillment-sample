#!/bin/bash

# Private qluvio Go modules are fetched over SSH (podman build --ssh default), like build-pushservice-container.sh; no token.
base_run_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
(
  cd ${HOME}/src
  _branch=${1:-main}
  _commit=$(git ls-remote https://github.com/eluv-io/redeemable-fulfillment-sample ${_branch} | head -1 | awk '{print $1}')

  podman build \
    --ssh default \
    --build-arg GIT_COMMIT=${_commit} \
    --build-arg GIT_BRANCH=${_branch} \
    -t fulfillmentd \
    -t us-docker.pkg.dev/eluvio-containers/containers/fulfillmentd:latest \
    -t us-docker.pkg.dev/eluvio-containers/containers/fulfillmentd:${_branch}-${_commit:0:8} \
    -f fulfillmentd.dockerfile \
    .

  if [ $? -eq 0 ]; then
    bash ${HOME}/auth-podman.sh || exit 1
    podman push \
      us-docker.pkg.dev/eluvio-containers/containers/fulfillmentd:latest
    podman push \
      us-docker.pkg.dev/eluvio-containers/containers/fulfillmentd:${_branch}-${_commit:0:8}
  fi
)
