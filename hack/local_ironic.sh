#!/bin/bash

set -xe

sudo podman run -d --net host --name mariadb --entrypoint /bin/runmariadb quay.io/metal3-io/ironic:master
sudo podman run -d --net host --name ironic -e "PROVISIONING_IP=127.0.0.1" quay.io/metal3-io/ironic:master
sudo podman run -d --net host --name ironic-inspector -e "PROVISIONING_IP=127.0.0.1" quay.io/metal3-io/ironic-inspector:master

for p in 3306 6385 5050; do
  nc -z -w 60 127.0.0.1 ${p}
done
