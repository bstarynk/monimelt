#!/bin/bash
# file post-merge-githook.sh to be installed as a git hook in .git/hooks/post-merge
# also invoked thru git pull
echo start post-merge-githook.sh "$@"
./monimelt-restore-state.sh  monimelt_state.sqlite monimelt_state.sql
echo end post-merge-githook.sh "$@"
