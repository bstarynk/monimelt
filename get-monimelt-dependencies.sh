#!/bin/sh
# file get-monimelt-dependencies.sh
if [ -z "$GOPATH" ]; then
    echo missing GOPATH, consider setting it e.g. with export GOPATH='$HOME/mygoworkspace' >&2
    exit 1
fi
grep url .gitmodules | sed 's+.*https://github.com/+go get -v github.com/+' | sh -x
