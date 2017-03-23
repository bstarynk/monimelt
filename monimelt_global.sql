-- monimelt_global.sql dump 2017 Mar 09 from monimelt_global.sqlite dumped by ./monimelt-dump-state.sh .....

 --   Copyright (C) 2017 Basile Starynkevitch.
 --  This sqlite3 dump file monimelt_global.sql is part of MONIMELT.
 --
 --  MONIMELT is free software; you can redistribute it and/or modify
 --  it under the terms of the GNU General Public License as published by
 --  the Free Software Foundation; either version 3, or (at your option)
 --  any later version.
 --
 --  MONIMELT is distributed in the hope that it will be useful,
 --  but WITHOUT ANY WARRANTY; without even the implied warranty of
 --  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 --  GNU General Public License for more details.
 --  You should have received a copy of the GNU General Public License
 --  along with MONIMELT; see the file COPYING3.   If not see
 --  <http://www.gnu.org/licenses/>.

PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE t_params 
 (par_name VARCHAR(35) PRIMARY KEY ASC NOT NULL UNIQUE, 
  par_value TEXT NOT NULL);
CREATE TABLE t_objects
 (ob_id VARCHAR(26) PRIMARY KEY ASC NOT NULL UNIQUE,
  ob_mtime INT NOT NULL,
  ob_jsoncont TEXT NOT NULL,
  ob_paylkind VARCHAR(40) NOT NULL,
  ob_paylcont TEXT NOT NULL);
INSERT INTO "t_objects" VALUES('_02hL3RuX4x6_6y6PTK9vZs7',1489052005,'
{
 "attrs": [],
 "comps": []
}
','','');
INSERT INTO "t_objects" VALUES('_1xKb8cfVIXo_7zufUqzNXfu',1489052005,'
{
 "attrs": [],
 "comps": []
}
','','');
CREATE TABLE t_globals
 (glob_name VARCHAR(80) PRIMARY KEY ASC NOT NULL UNIQUE,
  glob_oid VARCHAR(26)  NOT NULL);
COMMIT;
-- monimelt-dump-state end dump monimelt_global.sql
