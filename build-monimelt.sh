#!/bin/bash
# file build-monimelt.sh

## see https://blog.ksub.org/bytes/2017/02/12/exploring-shared-objects-in-go/
function build_our_package () {
    ourpkg=$1
    echo '=+=+= ' building $ourpkg
    time go tool compile -L -pack -linkshared -buildmode=shared $ourpkg/*go
    echo '=+=+= ' built $ourpkg
    echo
}

echo
echo '+*+*+*+*+' building our packages
echo

## did not work: time go install  -linkshared -buildmode=shared -v *mo

## order of our packages matter
build_our_package serialmo
build_our_package objvalmo
build_our_package payloadmo
echo
echo
echo '+*+*+*+*+' building the Monimelt program
time go build -linkshared -buildmode=exe -v monimelt
echo
echo
times
echo
