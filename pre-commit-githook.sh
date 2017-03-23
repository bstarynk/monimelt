#!/bin/bash
# file pre-commit-githook.sh to be installed as a git hook in .git/hooks/pre-commit
echo start pre-commit-githook.sh "$@"
./monimelt-dump-state.sh monimelt_global.sqlite monimelt_global.sql
echo end pre-commit-githook.sh "$@"
