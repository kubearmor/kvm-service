#!/bin/bash
set -e
set -x
shopt -s extglob

CLUSTER_PORT=38514
CLUSTER_IP=

if [[ $(which docker) && $(docker --version) ]]; then
    echo "Docker is installed!!!"
else
  echo "Failed: Docker is not installed!!!"
  exit -1
fi

export WORKLOAD_IDENTITY=38514

DOCKER_OPTS=" -d -p 32767:32767 --log-driver syslog --restart always"
DOCKER_OPTS+=" --privileged --add-host kvms.kubearmor.io:$CLUSTER_IP"
DOCKER_OPTS+=" --volume /var/run/docker.sock:/var/run/docker.sock"
DOCKER_OPTS+=" --volume /usr/src:/usr/src"
DOCKER_OPTS+=" --volume /lib/modules:/lib/modules"
DOCKER_OPTS+=" --volume /sys/fs/bpf:/sys/fs/bpf"
DOCKER_OPTS+=" --volume /sys/kernel/debug:/sys/kernel/debug"
DOCKER_OPTS+=" --volume /etc/apparmor.d:/etc/apparmor.d"
DOCKER_OPTS+=" --volume /etc/os-release:/media/root/etc/os-release"

KUBEARMOR_OPTS="-gRPC=32767 -logPath=/tmp/kubearmor.log -enableKubeArmorHostPolicy"

if [ -n "$(sudo docker ps -a -q -f name=kubearmor)" ]; then
    echo "Shutting down running kubearmor agent"
    sudo docker rm -f kubearmor || true
fi

KUBEARMOR_IMAGE="kubearmor/kubearmor:latest"

echo "Launching kubearmor agent..."
sudo docker run --name kubearmor $DOCKER_OPTS $KUBEARMOR_IMAGE kubearmor $KUBEARMOR_OPTS

