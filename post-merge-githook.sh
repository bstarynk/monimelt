#!/bin/bash
# file post-merge-githook.sh to be installed as a git hook in .git/hooks/post-merge
# also incoked thru git pull
echo start post-merge-githook.sh "$@"
make restorestate
echo end post-merge-githook.sh "$@"
