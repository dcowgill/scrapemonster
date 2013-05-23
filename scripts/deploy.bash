#!/bin/bash

set -eux

TARGET_HOST=scrapemonster@artanis.launchtime.com
RSYNC="rsync -avz"

GOARCH=amd64
GOOS=linux
GOPATH=$GOPATH # ensure variable is set

script_dir=$(cd `dirname "${BASH_SOURCE[0]}"` && pwd)
base_dir=`dirname $script_dir`

cd $base_dir && GOARCH=$GOARCH GOOS=$GOOS CGO_ENABLED=0 make
ssh $TARGET_HOST mkdir -p scrapemonster
$RSYNC $GOPATH/bin/${GOOS}_${GOARCH}/ $TARGET_HOST:scrapemonster/bin
$RSYNC $script_dir $TARGET_HOST:scrapemonster/
