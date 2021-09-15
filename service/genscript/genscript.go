// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var ewFile *os.File

func addContent(content string) {
    _, err := ewFile.WriteString(content + "\n")
	if err != nil {
        panic(err)
		log.Fatal(err)
	}
}

func main() {
	externalworkloadPtr := flag.String("external-workload", "", "External workload name/id")

	flag.Parse()
	if len(*externalworkloadPtr) == 0 {
		fmt.Println("Usage: command -external-workload")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ewFileName := "ew-" + *externalworkloadPtr + ".sh"

    var err error
	// Creating an empty file Using Create() function
    ewFile, err = os.Create(ewFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer ewFile.Close()

	err = os.Chmod(ewFileName, 0777)
	if err != nil {
		log.Fatal(err)
	}

    addContent("#!/bin/bash")
    addContent("set -e")
    addContent("set -x")
    addContent("shopt -s extglob")
    addContent("")

    addContent("if [ -z \"$CLUSTER_ADDR\" ] ; then")
    addContent("    echo \"CLUSTER_ADDR must be defined to the IP:PORT at which the KVM Service is reachable.\"")
    addContent("    exit 1")
    addContent("fi")
    addContent("")

    addContent("port='@(6553[0-5]|655[0-2][0-9]|65[0-4][0-9][0-9]|6[0-4][0-9][0-9][0-9]|[1-5][0-9][0-9][0-9][0-9]|[1-9][0-9][0-9][0-9]|[1-9][0-9][0-9]|[1-9][0-9]|[1-9])'")
    addContent("byte='@(25[0-5]|2[0-4][0-9]|[1][0-9][0-9]|[1-9][0-9]|[0-9])'")
    addContent("ipv4=\"$byte\\.$byte\\.$byte\\.$byte\"")
    addContent("")

    addContent("# Default port is for a HostPort service")
    addContent("case \"$CLUSTER_ADDR\" in")
    addContent("    \\[+([0-9a-fA-F:])\\]:$port)")
    addContent("    CLUSTER_PORT=${CLUSTER_ADDR##\\[*\\]:}")
    addContent("    CLUSTER_IP=${CLUSTER_ADDR#\\[}")
    addContent("    CLUSTER_IP=${CLUSTER_IP%\\]:*}")
    addContent("    ;;")
    addContent("    [^[]$ipv4:$port)")
    addContent("    CLUSTER_PORT=${CLUSTER_ADDR##*:}")
    addContent("    CLUSTER_IP=${CLUSTER_ADDR%:*}")
    addContent("    ;;")
    addContent("    *:*)")
    addContent("    echo \"Malformed CLUSTER_ADDR: $CLUSTER_ADDR\"")
    addContent("    exit 1")
    addContent("    ;;")
    addContent("    *)")
    addContent("    CLUSTER_PORT=12345")
    addContent("    CLUSTER_IP=$CLUSTER_ADDR")
    addContent("    ;;")
    addContent("esac")
    addContent("")

    addContent("DOCKER_OPTS=\" -d -p 32767:32767 --log-driver syslog --restart always\"")
    addContent("DOCKER_OPTS+=\" --privileged --add-host kvms.kubearmor.io:$CLUSTER_IP\"")
    addContent("DOCKER_OPTS+=\" --volume /var/run/docker.sock:/var/run/docker.sock\"")
    addContent("DOCKER_OPTS+=\" --volume /usr/src:/usr/src\"")
    addContent("DOCKER_OPTS+=\" --volume /lib/modules:/lib/modules\"")
    addContent("DOCKER_OPTS+=\" --volume /sys/fs/bpf:/sys/fs/bpf\"")
    addContent("DOCKER_OPTS+=\" --volume /sys/kernel/debug:/sys/kernel/debug\"")
    addContent("DOCKER_OPTS+=\" --volume /etc/apparmor.d:/etc/apparmor.d\"")
    addContent("DOCKER_OPTS+=\" --volume /etc/os-release:/media/root/etc/os-release\"")
    addContent("")
    addContent("KUBEARMOR_OPTS=\"-gRPC=32767 -logPath=/tmp/kubearmor.log -enableKubeArmorHostPolicy\"")
    addContent("")
    addContent("if [ -n \"$(sudo docker ps -a -q -f name=kubearmor)\" ]; then")
    addContent("    echo \"Shutting down running kubearmor agent\"")
    addContent("    sudo docker rm -f kubearmor || true")
    addContent("fi")
    addContent("")
    addContent("KUBEARMOR_IMAGE=\"kubearmor/kubearmor:latest\"")
    addContent("")
    addContent("echo \"Launching kubearmor agent...\"")
    addContent("sudo docker run --name kubearmor $DOCKER_OPTS $KUBEARMOR_IMAGE kubearmor $KUBEARMOR_OPTS")
    addContent("")

}
