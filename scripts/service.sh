#!/bin/bash

DOCKER_OPTS=" -d -p 32769:32769 --log-driver syslog --restart always"
DOCKER_OPTS+=" --privileged --add-host kvms.kvmservice.io:1.1.1.1"
DOCKER_OPTS+=" --volume /var/run/docker.sock:/var/run/docker.sock"
DOCKER_OPTS+=" --volume /usr/src:/usr/src"
DOCKER_OPTS+=" --volume /lib/modules:/lib/modules"
DOCKER_OPTS+=" --volume /sys/fs/bpf:/sys/fs/bpf"
DOCKER_OPTS+=" --volume /sys/kernel/debug:/sys/kernel/debug"
DOCKER_OPTS+=" --volume /etc/os-release:/media/root/etc/os-release"

kvmservice_OPTS="-port=32769 -ipAddress=1.1.1.1"

if [ -n "$(sudo docker ps -a -q -f name=kvmservice)" ]; then
    echo "Shutting down running kvmservice agent"
    sudo docker rm -f kvmservice || true
fi

kvmservice_IMAGE="quay.io/maddy007_maha/kvmservice:v3"

echo "Launching kvmservice agent..."
#sudo docker run --name kvmservice $DOCKER_OPTS $kvmservice_IMAGE kvmservice $kvmservice_OPTS
sudo docker run --name kvmservice -d -p 32769:32769 --log-driver syslog --restart always  --volume /var/run/docker.sock:/var/run/docker.sock --volume /etc/kubernetes/pki/etcd/:/etc/kubernetes/pki/etcd/ --privileged $kvmservice_IMAGE kvmservice $kvmservice_OPTS


