#!/bin/sh
# file build-monimelt.sh

## see https://blog.ksub.org/bytes/2017/02/12/exploring-shared-objects-in-go/
echo
echo '+*+*+*+*+' building our packages
time go install  -linkshared -buildmode=shared -v *mo
echo
echo
echo '+*+*+*+*+' building the Monimelt program
time go build -linkshared -buildmode=exe -v monimelt
echo
echo
times
echo
