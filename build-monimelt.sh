#!/bin/sh
# file build-monimelt.sh
if [ -z "$GOPATH" ]; then
    echo missing GOPATH, consider setting it e.g. with export GOPATH='$HOME/mygoworkspace' >&2
    exit 1
fi
if echo $GOPATH | grep $PWD ; then
   echo GOPATH "$GOPATH" contains current directory $PWD
else
    export GOPATH=$GOPATH:$PWD/
    echo updated GOPATH "$GOPATH"
fi
echo
echo building our packages
time go build -buildmode=shared -v serialmo objvalmo payloadmo
echo
echo
echo build the program
time go build -linkshared -buildmode=exe -v monimelt
echo
