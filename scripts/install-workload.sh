#!/bin/bash
set -e
#set -x
shopt -s extglob

if [ -z "$CLUSTER_ADDR" ] ; then
    echo "CLUSTER_ADDR must be defined to the IP:PORT at which the KVM Service is reachable."
    exit 1
fi

port='@(6553[0-5]|655[0-2][0-9]|65[0-4][0-9][0-9]|6[0-4][0-9][0-9][0-9]|[1-5][0-9][0-9][0-9][0-9]|[1-9][0-9][0-9][0-9]|[1-9][0-9][0-9]|[1-9][0-9]|[1-9])'
byte='@(25[0-5]|2[0-4][0-9]|[1][0-9][0-9]|[1-9][0-9]|[0-9])'
ipv4="$byte\.$byte\.$byte\.$byte"

# Default port is for a HostPort service
case "$CLUSTER_ADDR" in
    \[+([0-9a-fA-F:])\]:$port)
	CLUSTER_PORT=${CLUSTER_ADDR##\[*\]:}
	CLUSTER_IP=${CLUSTER_ADDR#\[}
	CLUSTER_IP=${CLUSTER_IP%\]:*}
	;;
    [^[]$ipv4:$port)
	CLUSTER_PORT=${CLUSTER_ADDR##*:}
	CLUSTER_IP=${CLUSTER_ADDR%:*}
	;;
    *:*)
	echo "Malformed CLUSTER_ADDR: $CLUSTER_ADDR"
	exit 1
	;;
    *)
	CLUSTER_PORT=12345
	CLUSTER_IP=$CLUSTER_ADDR
	;;
esac

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
