-- generated monimelt global dumpfile monimelt_state.sql
PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE t_params 
 (par_name VARCHAR(35) PRIMARY KEY ASC NOT NULL UNIQUE, 
  par_value TEXT NOT NULL);
CREATE TABLE t_objects
 (ob_id VARCHAR(26) PRIMARY KEY ASC NOT NULL UNIQUE,
  ob_mtime DATETIME,
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
-- end of monimelt global dumpfile monimelt_state.sql
