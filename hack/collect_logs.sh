#!/bin/bash

set -xe

if $(which podman > /dev/null 2>&1); then
    CONTAINER_RUNTIME=podman
else
    CONTAINER_RUNTIME=docker
fi

set +x

echo "******************** IRONIC LOGS ********************"
sudo $CONTAINER_RUNTIME logs ironic || true
echo
echo "******************* INSPECTOR LOGS ******************"
sudo $CONTAINER_RUNTIME logs ironic-inspector || true
