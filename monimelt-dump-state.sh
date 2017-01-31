#! /bin/sh
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

## Dont change the name monimelt-dump-state.sh of this script without
## care, it appears elsewhere (in Makefile & in monimelt.h)

echo start $0 "$@"
dbfile=$1
sqlfile=$2

if [ ! -f "$dbfile" ]; then
    echo "$0": missing database file "$dbfile" >& 2
    exit 1
fi

if file "$dbfile" | grep -qi SQLite ; then
    echo "$0:" dumping Monimelt Sqlite database $dbfile
else
    echo "$0:" bad database file "$dbfile" >& 2
    exit 1
fi

tempdump=$(basename $(tempfile -d . -p _tmp_ -s .sql))
trap 'rm -f $tempdump' EXIT INT QUIT TERM
export LANG=C LC_ALL=C

sqlbase=$(basename "$sqlfile")
dbbase=$(basename "$dbfile")
# generate an initial comment, it should be at least 128 bytes
date -r "$dbfile" +"-- $sqlbase dump %Y %b %d from $dbbase dumped by $0 ....." > $tempdump
echo >> $tempdump
date +' --   Copyright (C) %Y Basile Starynkevitch.' >> $tempdump
echo " --  This sqlite3 dump file $sqlbase is part of MONIMELT." >> $tempdump
echo ' --' >> $tempdump
echo ' --  MONIMELT is free software; you can redistribute it and/or modify' >> $tempdump
echo ' --  it under the terms of the GNU General Public License as published by' >> $tempdump
echo ' --  the Free Software Foundation; either version 3, or (at your option)' >> $tempdump
echo ' --  any later version.' >> $tempdump
echo ' --' >> $tempdump
echo ' --  MONIMELT is distributed in the hope that it will be useful,' >> $tempdump
echo ' --  but WITHOUT ANY WARRANTY; without even the implied warranty of' >> $tempdump
echo ' --  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the' >> $tempdump
echo ' --  GNU General Public License for more details.' >> $tempdump
echo ' --  You should have received a copy of the GNU General Public License' >> $tempdump
echo ' --  along with MONIMELT; see the file COPYING3.   If not see' >> $tempdump
echo ' --  <http://www.gnu.org/licenses/>.' >> $tempdump
echo >> $tempdump

sqlite3 $dbbase .dump >> $tempdump

echo "-- monimelt-dump-state end dump $dbbase" >> $tempdump



if [ -e "$sqlfile" ]; then
    # if only the first 128 bytes changed, it is some comment
    if cmp --quiet --ignore-initial 128 "$sqlfile" $tempdump ; then
	echo $0: unchanged Monimelt Sqlite3 dump "$sqlfile"
	exit 0
    fi
    echo -n "backup Monimelt Sqlite3 dump:" 
    mv -v --backup=existing "$sqlfile" "$sqlfile~"
fi

mv $tempdump "$sqlfile"

## we need that the .sql file has the same date as the .sqlite file
touch -f "$dbfile" "$sqlfile"
ls -l "$sqlfile"
