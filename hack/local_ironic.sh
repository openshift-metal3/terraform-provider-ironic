#!/bin/bash

set -xe

if $(which podman > /dev/null 2>&1); then
    CONTAINER_RUNTIME=podman
else
    CONTAINER_RUNTIME=docker
fi

IMAGE=${IMAGE:-quay.io/metal3-io/ironic:main}

sudo $CONTAINER_RUNTIME run -d --net host --privileged --name ironic \
    --entrypoint /bin/runironic -e "PROVISIONING_IP=127.0.0.1" $IMAGE
sudo $CONTAINER_RUNTIME run -d --net host --privileged --name ironic-inspector \
    --entrypoint /bin/runironic-inspector -e "PROVISIONING_IP=127.0.0.1" $IMAGE

for attempt in {1..30}; do
    sleep 2

    if ! curl -I http://127.0.0.1:6385; then
        if [[ $attempt == 30 ]]; then
            exit 1
        else
            continue
        fi
    fi

    if ! curl -I http://127.0.0.1:5050; then
        if [[ $attempt == 30 ]]; then
            exit 1
        else
            continue
        fi
    fi

    break
done
