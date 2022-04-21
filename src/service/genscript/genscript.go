// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package genscript

import (
	"strconv"

	kg "github.com/kubearmor/KVMService/src/log"
)

var (
	p          GenScriptParams
	ScriptData string
)

type GenScriptParams struct {
	port      uint16
	etcdPort  uint16
	ipAddress string
}

func InitGenScript(Port, etcdPort uint16, IpAddress string) {
	p.port = Port
	p.ipAddress = IpAddress
	p.etcdPort = etcdPort
}

func addContent(content string) {
	ScriptData = ScriptData + content + "\n"
}

func GenerateEWInstallationScript(virtualmachine, identity string) string {

	ScriptData = ""

	kg.Printf("Generating the installation script =>")
	kg.Printf("ClusterIP:%s ClusterPort:%d ewName:%s identity:%s", p.ipAddress, p.port, virtualmachine, identity)

	addContent("#!/bin/bash")
	addContent("set -e")
	addContent("shopt -s extglob")
	addContent("")

	contentStr := "CLUSTER_PORT=" + strconv.FormatUint(uint64(p.port), 10)
	addContent(contentStr)
	contentStr = "CLUSTER_IP=" + p.ipAddress
	addContent(contentStr)
	addContent("")

	addContent("if [[ $(which docker) && $(docker --version) ]]; then")
	addContent("    echo \"Docker is installed!!!\"")
	addContent("else")
	addContent("  echo \"Failed: Docker is not installed!!!\"")
	addContent("  exit -1")
	addContent("fi")
	addContent("")

	contentStr = "WORKLOAD_IDENTITY=" + identity
	addContent(contentStr)
	addContent("")

	addContent("DOCKER_OPTS=\" -d -p 32767:32767 --log-driver syslog --restart always\"")
	addContent("DOCKER_OPTS+=\" --privileged --add-host kvms.kubearmor.io:$CLUSTER_IP\"")
	addContent("DOCKER_OPTS+=\" --env CLUSTER_PORT=$CLUSTER_PORT --env CLUSTER_IP=$CLUSTER_IP\"")
	addContent("DOCKER_OPTS+=\" --env  WORKLOAD_IDENTITY=$WORKLOAD_IDENTITY\"")
	addContent("DOCKER_OPTS+=\" --volume /var/run/docker.sock:/var/run/docker.sock\"")
	addContent("DOCKER_OPTS+=\" --volume /usr/src:/usr/src\"")
	addContent("DOCKER_OPTS+=\" --volume /lib/modules:/lib/modules\"")
	addContent("DOCKER_OPTS+=\" --volume /sys/fs/bpf:/sys/fs/bpf\"")
	addContent("DOCKER_OPTS+=\" --volume /sys/kernel/debug:/sys/kernel/debug\"")
	addContent("DOCKER_OPTS+=\" --volume /etc/apparmor.d:/etc/apparmor.d\"")
	addContent("DOCKER_OPTS+=\" --volume /etc/os-release:/media/root/etc/os-release\"")
	addContent("")
	addContent("KUBEARMOR_OPTS=\" -enableKubeArmorVm -k8s=false -logPath=/tmp/kubearmor.log\"")
	addContent("")
	addContent("if [ -n \"$(sudo docker ps -a -q -f name=kubearmor)\" ]; then")
	addContent("    echo \"Shutting down running kubearmor agent\"")
	addContent("    sudo docker rm -f kubearmor || true")
	addContent("fi")
	addContent("")
	addContent("KUBEARMOR_IMAGE=\"kubearmor/kubearmor:stable\"")
	addContent("")
	addContent("echo \"Launching kubearmor agent...\"")
	addContent("sudo docker run --name kubearmor $DOCKER_OPTS $KUBEARMOR_IMAGE $KUBEARMOR_OPTS")
	addContent("")

	cilium := `
#####################################
#              CILIUM               #
#####################################
echo "Installing Cilium agent..."

`
	addContent(cilium)
	addContent("CILIUM_NODE=${CILIUM_NODE:-" + virtualmachine + "}")
	addContent("CILIUM_ETCD_PORT=${CILIUM_ETCD_PORT:-" + strconv.Itoa(int(p.etcdPort)) + "}")

	cilium = `
CILIUM_IMAGE=${CILIUM_IMAGE:-accuknox/cilium:latest}
HOST_IF=${HOST_IF:-eth+,en+}
CILIUM_CONFIG=${CILIUM_CONFIG:---devices=$HOST_IF --enable-host-firewall --enable-hubble=true --hubble-listen-address=:4244 --hubble-disable-tls=true --external-workload --enable-well-known-identities=false}

sudo mkdir -p /var/lib/cilium/etcd
sudo tee /var/lib/cilium/etcd/config.yaml <<EOF >/dev/null
---
insecure-transport: true
endpoints:
- http://$CLUSTER_IP:$CILIUM_ETCD_PORT
EOF

CILIUM_OPTS=" --join-cluster --enable-host-reachable-services --enable-endpoint-health-checking=false"
CILIUM_OPTS+=" --kvstore etcd --kvstore-opt etcd.config=/var/lib/cilium/etcd/config.yaml"
CILIUM_OPTS+=" $CILIUM_CONFIG"
if [ -n "$HOST_IP" ] ; then
    CILIUM_OPTS+=" --ipv4-node $HOST_IP"
fi

DOCKER_OPTS=" --env CEW_NAME=$CILIUM_NODE"
DOCKER_OPTS+=" -d --log-driver local --restart always"
DOCKER_OPTS+=" --privileged --network host --cap-add NET_ADMIN --cap-add SYS_MODULE"
DOCKER_OPTS+=" --cgroupns=host"
DOCKER_OPTS+=" --volume /var/lib/cilium/etcd:/var/lib/cilium/etcd"
DOCKER_OPTS+=" --volume /var/run/cilium:/var/run/cilium"
DOCKER_OPTS+=" --volume /boot:/boot"
DOCKER_OPTS+=" --volume /lib/modules:/lib/modules"
DOCKER_OPTS+=" --volume /sys/fs/bpf:/sys/fs/bpf"
DOCKER_OPTS+=" --volume /run/xtables.lock:/run/xtables.lock"

cilium_started=false
retries=4
while [ $cilium_started = false ]; do
    if [ -n "$(${SUDO} docker ps -a -q -f name=cilium)" ]; then
        echo "Shutting down running Cilium agent"
        ${SUDO} docker rm -f cilium || true
    fi

    echo "Launching Cilium agent $CILIUM_IMAGE..."
    ${SUDO} docker run --name cilium $DOCKER_OPTS $CILIUM_IMAGE cilium-agent $CILIUM_OPTS

    # Copy Cilium CLI
    ${SUDO} docker cp cilium:/usr/bin/cilium /usr/bin/cilium

    # Wait for cilium agent to become available
    for ((i = 0 ; i < 12; i++)); do
        if cilium status --brief > /dev/null 2>&1; then
            cilium_started=true
            break
        fi
        sleep 5s
        echo "Waiting for Cilium daemon to come up..."
    done

    echo "Cilium status:"
    cilium status || true

    if [ "$cilium_started" = true ] ; then
        echo 'Cilium successfully started!'
    else
        if [ $retries -eq 0 ]; then
            >&2 echo 'Timeout waiting for Cilium to start, retries exhausted.'
            exit 1
        fi
        ((retries--))
        echo "Restarting Cilium..."
    fi
done
`
	addContent(cilium)

	kg.Printf("Script data is successfully generated!")

	return ScriptData
}
