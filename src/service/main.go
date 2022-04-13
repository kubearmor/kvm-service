// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"

	kg "github.com/kubearmor/KVMService/src/log"
	cfg "github.com/kubearmor/KVMService/src/service/config"
	"github.com/kubearmor/KVMService/src/service/core"
)

// GitCommit string passed from govvv
var GitCommit string

// GitBranch string passed from govvv
var GitBranch string

// BuildDate string passed from govvv
var BuildDate string

// Version string passed from govvv
var Version string

func printBuildDetails() {
	kg.Printf("BUILD-INFO: commit:%s, branch: %s, date: %s, version: %s",
		GitCommit, GitBranch, BuildDate, Version)
}

func main() {
	printBuildDetails()
	if os.Geteuid() != 0 {
		kg.Printf("Need to have root privileges to run %s\n", os.Args[0])
		return
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		kg.Err(err.Error())
		return
	}

	if err := os.Chdir(dir); err != nil {
		kg.Err(err.Error())
		return
	}

	err = cfg.LoadConfig()
	if err != nil {
		kg.Err(err.Error())
		return
	}

	// == //

	core.KVMSDaemon()

	// == //
}
