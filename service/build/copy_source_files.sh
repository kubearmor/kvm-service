#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright 2021 Authors of KVMService

realpath() {
    CURR=$PWD

    cd "$(dirname "$0")"
    LINK=$(readlink "$(basename "$0")")

    while [ "$LINK" ]; do
        cd "$(dirname "$LINK")"
        LINK=$(readlink "$(basename "$1")")
    done

    REALPATH="$PWD/$(basename "$1")"
    echo "$REALPATH"

    cd $CURR
}

ARMOR_HOME=`dirname $(realpath "$0")`/..
SERVICE_PATH=`dirname $(realpath "$0")`/..

mkdir -p $ARMOR_HOME/build/KVMService

# copy files to build
cp -r $SERVICE_PATH/common/ $ARMOR_HOME/build/KVMService/
cp -r $SERVICE_PATH/core/ $ARMOR_HOME/build/KVMService/
cp -r $SERVICE_PATH/log/ $ARMOR_HOME/build/KVMService/
cp -r $SERVICE_PATH/types/ $ARMOR_HOME/build/KVMService/
cp -r $SERVICE_PATH/etcd/ $ARMOR_HOME/build/KVMService/
cp -r $SERVICE_PATH/constants/ $ARMOR_HOME/build/KVMService/
cp -r $SERVICE_PATH/server/ $ARMOR_HOME/build/KVMService/
cp -r $SERVICE_PATH/protobuf/ $ARMOR_HOME/build/KVMService/
cp $SERVICE_PATH/go.mod $ARMOR_HOME/build/KVMService/
cp $SERVICE_PATH/main.go $ARMOR_HOME/build/KVMService/

 cp $SERVICE_PATH/build/patch.sh $ARMOR_HOME/build/KVMService/
 cp $SERVICE_PATH/build/patch_selinux.sh $ARMOR_HOME/build/KVMService/
