// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package genscript

import (
	//"flag"
	//"fmt"
	kg "github.com/kubearmor/KVMService/operator/log"
	"os"
	"strconv"
)

var (
	ewFile     *os.File
	p          GenScriptParams
	ScriptData string
)

type GenScriptParams struct {
	port      uint16
	ipAddress string
}

func InitGenScript(Port uint16, IpAddress string) {
	p.port = Port
	p.ipAddress = IpAddress
}

func addContent(content string) {
	ScriptData = ScriptData + content + "\n"
}

func GenerateEWInstallationScript(externalWorkload, identity string) string {

	kg.Printf("Generating the installation script with following args: =>")
	kg.Printf("ClusterIP:%s ClusterPort:%d ewName:%s identity:%s", p.ipAddress, p.port, externalWorkload, identity)

	/*
		ewFileName := "/mnt/gen-script/" + "ew-" + externalWorkload + ".sh"

		var err error
		// Creating an empty file Using Create() function
		ewFile, err = os.Create(ewFileName)
		if err != nil {
			kg.Printf("Error: File creation file:%s", ewFileName)
			log.Fatal(err)
		}
		defer ewFile.Close()

		err = os.Chmod(ewFileName, 0777)
		if err != nil {
			kg.Printf("File permissions failed file:%s", ewFileName)
			log.Fatal(err)
		}*/

	addContent("#!/bin/bash")
	addContent("set -e")
	addContent("set -x")
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

	//contentStr = "WORKLOAD_IDENTITY=" + strconv.FormatUint(uint64(identity), 10)
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
	addContent("KUBEARMOR_OPTS=\"-gRPC=32767 -logPath=/tmp/kubearmor.log -enableKubeArmorHostPolicy=true\"")
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

	addContent("DOCKER_USER=\"accuknox-user\"")
	addContent("DOCKER_PASS=\"5U:u8~wu-b\"")

	addContent("IP_ADDR=`hostname -I | cut --delimiter \" \" --fields 1`")

	addContent("FEEDER_IMG_PATH=\" agents.accuknox.com/repository/docker-dev/feeder-service:1.0.371\"")

	addContent("sudo docker login agents.accuknox.com -u $DOCKER_USER -p $DOCKER_PASS")

	addContent("sudo docker pull $FEEDER_IMG_PATH")

	addContent("FEEDER_OPTS=\" -d --log-driver syslog --restart always\"")
	addContent("FEEDER_OPTS+=\" --privileged --add-host kubearmor-dev:$CLUSTER_IP\"")
	addContent("FEEDER_OPTS+=\" --env KUBEARMOR_ENABLED=true\"")
	addContent("FEEDER_OPTS+=\" --env KUBEARMOR_URL=$IP_ADDR\"")
	addContent("FEEDER_OPTS+=\" --env KUBEARMOR_PORT=32767\"")
	addContent("FEEDER_OPTS+=\" --env KAFKA_ENABLED=true\"")
	addContent("FEEDER_OPTS+=\" --env KAFKA_URL=$CLUSTER_IP\"")
	addContent("FEEDER_OPTS+=\" --env KAFKA_PORT=9092\"")
	addContent("FEEDER_OPTS+=\" --env KAFKA_BOOTSTRAP_SERVERS=$CLUSTER_IP:9092\"")
	addContent("FEEDER_OPTS+=\" --env KAFKA_HOSTNAME=$CLUSTER_IP\"")

	addContent("sudo docker run --name feeder-service $FEEDER_OPTS $FEEDER_IMG_PATH ./bin/feeder-service")

	kg.Printf("Script data is successfully generated!")

	return ScriptData
}
