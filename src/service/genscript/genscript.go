// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package genscript

import (
	"io/ioutil"
	"strconv"

	ct "github.com/kubearmor/KVMService/src/constants"
	kg "github.com/kubearmor/KVMService/src/log"
)

var (
	p          GenScriptParams
	ScriptData string
)

type GenScriptParams struct {
	port      uint16
	ipAddress string
	CaCert    string
}

func InitGenScript(Port uint16, IpAddress string) {
	p.port = Port
	p.ipAddress = IpAddress

	byteData, err := ioutil.ReadFile(ct.CaCertPath)
	if err != nil {
		kg.Err(err.Error())
	}
	p.CaCert = "\"" + string(byteData) + "\""
}

func addContent(content string) {
	ScriptData = ScriptData + content + "\n"
}

func GenerateVMInstallationScript(virtualMachine string, identity string) string {

	ScriptData = ""

	kg.Printf("Generating the installation script =>")
	kg.Printf("ClusterIP:%s ClusterPort:%d vmName:%s identity:%s", p.ipAddress, p.port, virtualMachine, identity)

	addContent("#!/bin/bash")
	addContent("set -e")
	addContent("set -x")
	addContent("shopt -s extglob")
	addContent("")

	contentStr := "CLUSTER_PORT=" + strconv.FormatUint(uint64(p.port), 10)
	addContent(contentStr)
	contentStr = "CLUSTER_IP=" + p.ipAddress
	addContent(contentStr)
	addContent("CA_CERT_FILE_PATH_LOCAL=\"/home/$USER/ca.cert\"")
	addContent("CA_CERT_FILE_PATH_DEST=\"/media/root/ca.cert\"")
	addContent("")
	contentStr = "CA_CERT=" + p.CaCert
	addContent(contentStr)
	addContent("")
	addContent("echo \"$CA_CERT\" > $CA_CERT_FILE_PATH_LOCAL")
	addContent("")
	contentStr = "CA_CERT=" + p.CaCert
	addContent(contentStr)
	addContent("")

	addContent("if [[ $(which docker) && $(docker --version) ]]; then")
	addContent("    echo \"Docker is installed!!!\"")
	addContent("else")
	addContent("  echo \"Failed: Docker is not installed!!!\"")
	addContent("  exit -1")
	addContent("fi")
	addContent("")

	contentStr = "VM_IDENTITY=" + identity
	addContent(contentStr)
	addContent("")

	addContent("DOCKER_OPTS=\" -d -p 32767:32767 --log-driver syslog --restart always\"")
	addContent("DOCKER_OPTS+=\" --privileged --add-host kvms.kubearmor.io:$CLUSTER_IP\"")
	addContent("DOCKER_OPTS+=\" --env CLUSTER_PORT=$CLUSTER_PORT --env CLUSTER_IP=$CLUSTER_IP\"")
	addContent("DOCKER_OPTS+=\" --env VM_IDENTITY=$VM_IDENTITY\"")
	addContent("DOCKER_OPTS+=\" --env CA_CERT_FILE_PATH=$CA_CERT_FILE_PATH_DEST\"")
	addContent("DOCKER_OPTS+=\" --volume /var/run/docker.sock:/var/run/docker.sock\"")
	addContent("DOCKER_OPTS+=\" --volume /usr/src:/usr/src\"")
	addContent("DOCKER_OPTS+=\" --volume /lib/modules:/lib/modules\"")
	addContent("DOCKER_OPTS+=\" --volume /sys/fs/bpf:/sys/fs/bpf\"")
	addContent("DOCKER_OPTS+=\" --volume /sys/kernel/debug:/sys/kernel/debug\"")
	addContent("DOCKER_OPTS+=\" --volume /etc/apparmor.d:/etc/apparmor.d\"")
	addContent("DOCKER_OPTS+=\" --volume /etc/os-release:/media/root/etc/os-release\"")
	addContent("DOCKER_OPTS+=\" --volume $CA_CERT_FILE_PATH_LOCAL:CA_CERT_FILE_PATH_DEST\"")
	addContent("")
	addContent("KUBEARMOR_OPTS=\" -enableKubeArmorVm true -logPath=/tmp/kubearmor.log\"")
	addContent("")
	addContent("if [ -n \"$(sudo docker ps -a -q -f name=kubearmor)\" ]; then")
	addContent("    echo \"Shutting down running kubearmor agent\"")
	addContent("    sudo docker rm -f kubearmor || true")
	addContent("fi")
	addContent("")
	addContent("KUBEARMOR_IMAGE=\"kubearmor/kubearmor:latest\"")
	addContent("")
	addContent("echo \"Launching kubearmor agent...\"")
	addContent("sudo docker run --name kubearmor $DOCKER_OPTS $KUBEARMOR_IMAGE $KUBEARMOR_OPTS")
	addContent("")

	kg.Printf("Script data is successfully generated!")

	return ScriptData
}
