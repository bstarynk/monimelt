// file objvalmo/persistmo.go

/// use https://github.com/gwenn/gosqlite
package objvalmo

import (
	//"bytes"
	"database/sql"
	"fmt"
	gosqlite "github.com/gwenn/gosqlite"
	"log"
	"serialmo"
	"strings"
)

const GlobalObjects = true
const UserObjects = false

func sqliteerrorlogmo(d interface{}, err error, msg string) {
	log.Printf("SQLITE: %s, %s\n", err, msg)
}

func init() {
	err := gosqlite.ConfigLog(sqliteerrorlogmo, nil)
	if err == nil {
		panic(fmt.Errorf("persistmo could not ConfigLog sqlite: %v", err))
	}
}

type LoaderMo struct {
	ldglobaldb *sql.DB
	lduserdb   *sql.DB
	ldobjmap   *map[serialmo.IdentMo]*ObjectMo
}

func validsqlitepath(path string) bool {
	// check that path has no :?&=$~;' characters
	return !strings.ContainsAny(path, ":?&=$;'")
}

func OpenLoader(globalpath string, userpath string) *LoaderMo {
	if !validsqlitepath(globalpath) {
		panic(fmt.Errorf("objvalmo.OpenLoad invalid global path %s",
			globalpath))
	}
	l := new(LoaderMo)
	db, err := sql.Open("sqlite3", "file:"+globalpath+"?mode=readonly")
	if err != nil {
		panic(fmt.Errorf("objvalmo.OpenLoad failed to open global db %s - %v",
			globalpath, err))
	}
	l.ldglobaldb = db
	if len(userpath) > 0 {
		if !validsqlitepath(userpath) {
			panic(fmt.Errorf("objvalmo.OpenLoad invalid user path %s",
				userpath))
		}
		db, err := sql.Open("sqlite3", "file:"+userpath+"?mode=readonly")
		if err != nil {
			panic(fmt.Errorf("persistmo.OpenLoad failed to open user db %s - %v",
				userpath, err))
		}
		l.lduserdb = db
	}
	l.ldobjmap = new(map[serialmo.IdentMo]*ObjectMo)
	return l
}

func (l *LoaderMo) create_objects(globflag bool) {
	var qr *sql.Rows
	var err error
	if globflag {
		qr, err = l.ldglobaldb.Query("SELECT ob_id FROM to_objects")
	} else {
		qr, err = l.lduserdb.Query("SELECT ob_id FROM to_objects")
	}
	if err != nil {
		panic(fmt.Errorf("loader: create_objects failure %v", err))
	}
	defer qr.Close()
	for qr.Next() {
		var idstr string
		err = qr.Scan(&idstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.create_objects failure %v", err))
		}
		oid, err := serialmo.IdFromString(idstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.create_objects bad id %s: %v", idstr, err))
		}
		pob := MakeObjectById(oid)
		(*l.ldobjmap)[oid] = pob
	}
	err = qr.Err()
	if err != nil {
		panic(fmt.Errorf("persistmo.create_objects final %v", err))
	}
}

func (l *LoaderMo) Load() {
	l.create_objects(GlobalObjects)
	if l.lduserdb != nil {
		l.create_objects(UserObjects)
	}
}

func (l *LoaderMo) Close() {
	if l == nil {
		return
	}
	if ud := l.lduserdb; ud != nil {
		l.lduserdb = nil
		ud.Close()
	}
	if gd := l.ldglobaldb; gd != nil {
		l.ldglobaldb = nil
		gd.Close()
	}
	/// clear the object map
	l.ldobjmap = nil
}
