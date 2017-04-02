#!/bin/bash
# file get-monimelt-dependencies.sh
date +"start of get-monimelt-dependencies.sh at %c"

mygoarch=$(go env GOARCH)  ## e.g. amd64
mygoos=$(go env GOOS)      ## e.g. linux
mygoroot=$(go env GOROOT)  ## e.g. /usr/local/go


## check that go install -buildmode=shared std has been run once
if [ -f $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/fmt.shlibname ]; then
    myfmtshlibname=$(head -1 $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/fmt.shlibname)
    if [ ! -f $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/$myfmtshlibname ]; then
	echo 1>&2
	echo you should have done: go install -buildmode=shared std 1>&2
	echo ... but $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/$myfmtshlibname is missing 1>&2
	echo ... using content of $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/$myfmtshlibname
	echo >&2
	exit 1
    else
	echo you did run: go install -buildmode=shared std
	echo ... because we got: $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/fmt.shlibname
	echo ... pointing to: $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/$myfmtshlibname
    fi
    
else
    echo 1>&2
    echo you should have done: go install -buildmode=shared std 1>&2
    echo ... but $mygoroot/pkg/${mygoos}_${mygoarch}_dynlink/fmt.shlibname is missing 1>&2
    echo >&2
    exit 1
fi


function get_dependency () {
    echo '=*=*=*=*= +++' getting "$@"
    go get -u -v -buildmode=shared -linkshared $*
    failcod=$?
    if [ $failcod -gt 0 ]; then
	echo '!!!!!' failed to get "$@" : $failcod
	exit $failcod
    fi
    echo '=*=*=*=*= ---' got from "$@"
    echo    
}

get_dependency github.com/antonholmquist/jason
get_dependency -tags "libsqlite3 $mygoos" github.com/bstarynk/go-sqlite3
get_dependency github.com/ocdogan/rbt


date +"end of get-monimelt-dependencies.sh at %c%n"
