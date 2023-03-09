#!/bin/bash

set -xe

if $(which podman > /dev/null 2>&1); then
    CONTAINER_RUNTIME=podman
else
    CONTAINER_RUNTIME=docker
fi

sudo $CONTAINER_RUNTIME run -d --net host --privileged --name ironic \
    --entrypoint /bin/runironic -e "PROVISIONING_IP=127.0.0.1" quay.io/metal3-io/ironic:master
sudo $CONTAINER_RUNTIME run -d --net host --privileged --name ironic-inspector \
    -e "PROVISIONING_IP=127.0.0.1" quay.io/metal3-io/ironic-inspector:master

for p in 6385 5050; do
  nc -z -w 60 127.0.0.1 ${p}
done
