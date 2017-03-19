#!/bin/sh
# file build-monimelt.sh
if [ -z "$GOPATH" ]; then
    echo missing GOPATH, consider setting it e.g. with export GOPATH='$HOME/mygoworkspace' >&2
    exit 1
fi
if echo $GOPATH | grep $PWD ; then
   echo GOPATH contains current directory $PWD
else
    export GOPATH=$GOPATH:$PWD/
fi
go build -buildmode=exe -v monimelt
