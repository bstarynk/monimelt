// file objvalmo/persistmo.go

package objvalmo

import (
	"bytes"
	"fmt"
	"database/sql"
	_ "go-sqlite3"
)

/**
 #cgo pkg-config: sqlite3
 #include <stdio.h>
 #include <stdlib.h>
 #include <sqlite/sqlite.h>
 #include <pthread.h>
 void monimelt_sqlite3_errorlog(void *pdata __attribute__((unused)), int errcode, const char *msg) {
    fprintf(stderr, "monimelt Sqlite3 error:  errcode#%d msg=%s\n", errcode, msg);
    fflush(stderr);
 }
 void monimelt_initialize_sqlite3_logging(void) {
  sqlite3_config (SQLITE_CONFIG_LOG, monimelt_sqlite3_errorlog, NULL);
 }
**/

import "C"

func init() {
	C.monimelt_initialize_sqlite3_logging()
}

type LoaderMo struct {
	ldglobaldb *sql.DB
	lduserdb *sql.DB
	ldobjmap map[serialmo.IdentMo] *ObjectMo
}

func OpenLoader(globalpath string, userpath string) *LoaderMo {
	l := new(LoaderMo)
	l.ldglobaldb, err := sql.Open("sqlite3", globalpath)
	if err != nil {
		panic(fmt.Errorf("objvalmo.OpenLoad failed to open global db %s - %v",
			globalpath, err))
	}
	if (len(userpath) > 0) {
		l.lduserdb, err := sql.Open("sqlite3", userpath)
		if err != nil {
		panic(fmt.Errorf("persistmo.OpenLoad failed to open user db %s - %v",
			userpath, err))
		}
	}
	l.ldobjmap = new(map[serialmo.IdentMo] *ObjectMo)
	return l
}


func (l *LoaderMo) create_objects (globflag bool) {
	var qr *sql.Rows
	var err error
	if globflag {
		qr, err := l.ldglobaldb.Query("SELECT ob_id FROM to_objects")
	} else {
		qr, err := l.lduserdb.Query("SELECT ob_id FROM to_objects")
	}
	if err != nil {
		panic(fmt.Errorf("loader: create_objects failure %v"), err)
	}
	defer qr.Close()
	for qr.Next() {
		var idstr
		err = qr.Scan(&idstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.create_objects failure %v", err))
		}
		oid, err := serialmo.IdFromString(idstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.create_objects bad id %s: %v", idstr, err))
		}
		pob := objvalmo.MakeObjectById(oid)
		l.ldobjmap[oid] = pob
	}
	err = qr.Err()
	if err != nil {
		panic(fmt.Errorf("persistmo.create_objects final %v", err))
	}
}


func (l *LoaderMo) Load() {
	l.create_objects(true)
	if l.lduserdb {
		l.create_objects(false)
	}	
}

func (l *LoaderMo) Close() {
	if l == nil {
		return
	}
	if ud := l.lduserdb ; ud != nil {
		l.lduserdb = nil
		ud.Close()
	}
	if gd := l.ldglobaldb ; gd != nil {
		l.ldglobaldb = nil
		gd.Close()
	}
	/// clear the object map
	l.ldobjmap = new(map[serialmo.IdentMo] *ObjectMo)
}
