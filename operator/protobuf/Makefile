# SPDX-License-Identifier: Apache-2.0
# Copyright 2021 Authors of KubeArmor

CURDIR=$(shell pwd)

.PHONY: build
build:
	cd $(CURDIR); go get .
	cd $(CURDIR); protoc --proto_path=. --go_opt=paths=source_relative --go_out=plugins=grpc:. clihandler.proto
	cd $(CURDIR); protoc -I=. --go_out . --go_opt paths=source_relative --go-grpc_out . --go-grpc_opt paths=source_relative clihandler.proto


.PHONY: clean
clean:
	cd $(CURDIR); find . -name go.sum | xargs -I {} rm -f {}
