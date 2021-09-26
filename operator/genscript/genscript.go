// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package genscript

import (
	//"flag"
	//"fmt"
	"log"
	"os"
    "strconv"
)

var ewFile *os.File

func addContent(content string) {
	_, err := ewFile.WriteString(content + "\n")
	if err != nil {
		panic(err)
	}
}

func GenerateEWInstallationScript(port uint16, ipAddress, externalWorkload string, identity uint16) {
	/*
		externalworkloadPtr := flag.String("external-workload", "", "External workload name/id")

		flag.Parse()
		if len(*externalworkloadPtr) == 0 {
			fmt.Println("Usage: command -external-workload")
			flag.PrintDefaults()
			os.Exit(1)
		}
		ewFileName := "ew-" + *externalworkloadPtr + ".sh"
	*/

	ewFileName := "ew-" + externalWorkload + ".sh"

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

    contentStr := "CLUSTER_PORT=" + strconv.FormatUint(uint64(identity), 10)
	addContent(contentStr)
    contentStr = "CLUSTER_IP=" +  ipAddress
	addContent(contentStr)
	addContent("")

	addContent("if [[ $(which docker) && $(docker --version) ]]; then")
	addContent("    echo \"Docker is installed!!!\"")
	addContent("else")
    addContent("  echo \"Failed: Docker is not installed!!!\"")
	addContent("  exit -1")
	addContent("fi")
	addContent("")

	contentStr = "export WORKLOAD_IDENTITY=" + strconv.FormatUint(uint64(identity), 10)
	addContent(contentStr)
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
