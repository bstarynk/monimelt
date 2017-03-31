#!/bin/bash
# file get-monimelt-dependencies.sh
date +"start of get-monimelt-dependencies.sh at %c"

function get_github_dependency () {
    gd=$1
    echo '=*=*=*=*= +++' getting from github $gd
    go get -u -v -buildmode=shared -linkshared github.com/$gd
    echo '=*=*=*=*= ---' got from github $gd
    echo    
}

echo
echo you should have done: go install -buildmode=shared std
echo

get_github_dependency antonholmquist/jason
get_github_dependency mattn/go-sqlite3
get_github_dependency ocdogan/rbt


date +"end of get-monimelt-dependencies.sh at %c%n"
