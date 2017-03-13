#!/bin/sh

##   Copyright (C) 2017  Basile Starynkevitch
##   This file is part of MONIMELT.
## 
##   MONIMELT is free software; you can redistribute it and/or modify
##   it under the terms of the GNU General Public License as published by
##   the Free Software Foundation; either version 3, or (at your option)
##   any later version.
## 
##   MONIMELT is distributed in the hope that it will be useful,
##   but WITHOUT ANY WARRANTY; without even the implied warranty of
##   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
##   GNU General Public License for more details.
##   You should have received a copy of the GNU General Public License
##   along with MONIMELT; see the file COPYING3.   If not see
##   <http://www.gnu.org/licenses/>.

## Dont change the name monimelt-restore-state.sh of this script without
## care, it appears elsewhere

echo start $0 "$@"
dbfile=$1
sqlfile=$2

if [ -z "$dbfile" ] ; then
    dbfile=monimelt_state.sqlite
    sqlfile=monimelt_state.sql
fi

if [ -f "$dbfile" ] ; then
    mv -v --backup "$dbfile" "$dbfile~"
fi
if [ ! -f "$sqlfile" ]; then
	echo missing SQL file "$sqlfile"
	exit 1
fi
sqlite3 "$dbfile" <  "$sqlfile"
touch -r "$sqlfile" "$dbfile"
