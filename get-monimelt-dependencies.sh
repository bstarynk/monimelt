#!/bin/bash
# file get-monimelt-dependencies.sh
date +"start of get-monimelt-dependencies.sh at %c"
github_dependencies=( \
		      antonholmquist/jason \
			  mattn/go-sqlite3 \
			  ocdogan/rbt \
    )

for gd in $github_dependencies; do
    echo
    echo '=*=*=*=*= +++' getting from github $gd
    go get -u -v -buildmode=shared -linkshared github.com/$gd
    echo '=*=*=*=*= ---' got from github $gd
    echo
done

date +"end of get-monimelt-dependencies.sh at %c"
