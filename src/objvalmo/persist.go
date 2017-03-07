// file objvalmo/persist.go

package objvalmo

import (
	"database/sql"
	"fmt"
	gosqlite "github.com/gwenn/gosqlite"
	"log"
	"os"
	"serialmo"
	"strings"
)

const DefaultGlobalDbname = "monimelt_state"
const DefaultUserDbname = "monimelt_user"

const GlobalObjects = true
const UserObjects = false

func sqliteerrorlogmo(d interface{}, err error, msg string) {
	log.Printf("SQLITE: %s, %s\n", err, msg)
}

func init() {
	err := gosqlite.ConfigLog(sqliteerrorlogmo, nil)
	if err != nil {
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

func OpenLoaderFromFiles(globalpath string, userpath string) *LoaderMo {
	if !validsqlitepath(globalpath) {
		panic(fmt.Errorf("OpenLoaderFromFiles invalid global path %s",
			globalpath))
	}
	if _, err := os.Stat(globalpath); err != nil {
		panic(fmt.Errorf("OpenLoaderFromFiles wrong global path %s - %v",
			globalpath, err))
	}
	if len(userpath) > 0 {
		if !validsqlitepath(userpath) {
			panic(fmt.Errorf("OpenLoaderFromFiles invalid user path %s",
				userpath))
		}
		if _, err := os.Stat(userpath); err != nil {
			panic(fmt.Errorf("OpenLoaderFromFiles wrong user path %s - %v",
				userpath, err))
		}
	}
	l := new(LoaderMo)
	db, err := sql.Open("sqlite3", "file:"+globalpath+"?mode=ro")
	if err != nil {
		panic(fmt.Errorf("OpenLoaderFromFiles failed to open global db %s - %v",
			globalpath, err))
	}
	l.ldglobaldb = db
	if len(userpath) > 0 {
		db, err := sql.Open("sqlite3", "file:"+userpath+"?mode=ro")
		if err != nil {
			panic(fmt.Errorf("persistmo.OpenLoaderFromFiles failed to open user db %s - %v",
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
		qr, err = l.ldglobaldb.Query("SELECT ob_id FROM t_objects")
	} else {
		qr, err = l.lduserdb.Query("SELECT ob_id FROM t_objects")
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
	if l == nil {
		return
	}
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

const dump_chunk_len = 7

type dumpChunk struct {
	dchobjects [dump_chunk_len]*ObjectMo
	dchnext    *dumpChunk
}

const (
	dumod_Idle = iota
	dumod_Scan
	dumod_Emit
)

type DumperMo struct {
	dumode       uint
	dudirname    string
	dutempsuffix string
	duglobaldb   *sql.DB
	duuserdb     *sql.DB
	dufirstchk   *dumpChunk
	dulastchk    *dumpChunk
	dusetobjects *map[*ObjectMo]*ObjectMo
}

const sql_create_t_params = `CREATE TABLE IF NOT EXISTS t_params
 (par_name VARCHAR(35) PRIMARY KEY ASC NOT NULL UNIQUE,  par_value TEXT NOT NULL)`

const sql_create_t_objects = `CREATE TABLE IF NOT EXISTS t_objects
 (ob_id VARCHAR(26) PRIMARY KEY ASC NOT NULL UNIQUE,
  ob_mtime DATETIME,
  ob_jsoncont TEXT NOT NULL,
  ob_paylkind VARCHAR(40) NOT NULL,
  ob_paylcont TEXT NOT NULL)`

func (du DumperMo) create_tables(globflag bool) {
	var db *sql.DB
	if globflag {
		db = du.duglobaldb
	} else {
		db = du.duuserdb
	}
	if db == nil {
		panic(fmt.Errorf("create_tables no db in directory %s", du.dudirname))
	}
	_, err := db.Exec(sql_create_t_params)
	if err != nil {
		panic(fmt.Errorf("create_tables failure in directory %s for t_params creation %v",
			du.dudirname, err))
	}
	_, err = db.Exec(sql_create_t_objects)
	if err != nil {
		panic(fmt.Errorf("create_tables failure in directory %s for t_objects creation %v",
			du.dudirname, err))
	}
}

func (du DumperMo) AddDumpedObject(pob *ObjectMo) {
	if du.dumode != dumod_Scan {
		panic("AddDumpedObject in non-scanning dumper")
	}
	if pob == nil || pob.SpaceNum() == SpaTransient {
		return
	}
	if _, found := (*du.dusetobjects)[pob]; found {
		return
	}
	if du.dufirstchk == nil {
		nchk := new(dumpChunk)
		du.dufirstchk = nchk
		du.dulastchk = nchk
	}
	lchk := du.dulastchk
	putix := -1
	for ix := 0; ix < dump_chunk_len; ix++ {
		if lchk.dchobjects[ix] == nil {
			putix = ix
			break
		}
	}
	if putix >= 0 {
		lchk.dchobjects[putix] = pob
		return
	}
	nchk := new(dumpChunk)
	lchk.dchnext = nchk
	du.dulastchk = nchk
	nchk.dchobjects[0] = pob
	du.dulastchk = nchk
}

func OpenDumperDirectory(dirpath string) *DumperMo {
	if !validsqlitepath(dirpath) {
		panic(fmt.Errorf("OpenDumperDirectory invalid dirpath %q", dirpath))
	}
	if dirpath == "" {
		dirpath = "."
	}
	di, err := os.Stat(dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dirpath, 0750)
			if err != nil {
				panic(fmt.Errorf("OpenDumperDirectory cannot make dir %s : %v", dirpath, err))
			}
		} else {
			panic(fmt.Errorf("OpenDumperDirectory bad dirpath %s : %v", dirpath, err))
		}
	} else if !di.Mode().IsDir() {
		panic(fmt.Errorf("OpenDumperDirectory dirpath %s is not a directory", dirpath))
	}
	dtempsuf := fmt.Sprintf("+%v_p%d.tmp", serialmo.RandomSerial(), os.Getpid())
	globtemppath := fmt.Sprintf("%s/%s%s", dirpath, DefaultGlobalDbname, dtempsuf)
	usertemppath := fmt.Sprintf("%s/%s%s", dirpath, DefaultUserDbname, dtempsuf)
	glodb, err := sql.Open("sqlite3", "file:"+globtemppath+"?mode=rwc&cache=private")
	if err != nil {
		panic(fmt.Errorf("OpenDumperDirectory failed to open global db %s - %v", globtemppath, err))
	}
	usrdb, err := sql.Open("sqlite3", "file:"+usertemppath+"?mode=rwc&cache=private")
	if err != nil {
		glodb.Close()
		os.Remove(globtemppath)
		panic(fmt.Errorf("OpenDumperDirectory failed to open user db %s - %v", usertemppath, err))
	}
	du := new(DumperMo)
	du.dudirname = dirpath
	du.dutempsuffix = dtempsuf
	du.duglobaldb = glodb
	du.duuserdb = usrdb
	du.create_tables(GlobalObjects)
	du.create_tables(UserObjects)
	return du
}

func (du *DumperMo) StartDumpScan() {
	if du == nil || du.dumode != dumod_Idle {
		panic("StartDumpScan on non-idle dumper")
	}
	du.dumode = dumod_Scan
	DumpScanPredefined(du)
	DumpScanGlobalVariables(du)
}

func (du *DumperMo) IsDumpedObject(pob *ObjectMo) bool {
	_, found := (*du.dusetobjects)[pob]
	return found
}

func (du *DumperMo) LoopDumpScan() {
	if du == nil || du.dumode != dumod_Scan {
		panic("LoopDumpScan on non-scanning dumper")
	}
	var chk *dumpChunk
	var nchk *dumpChunk
	for chk = du.dufirstchk; chk != nil; chk = nchk {
		nchk = chk.dchnext
		if chk == du.dulastchk {
			du.dufirstchk = nil
			du.dulastchk = nil
		} else {
			du.dufirstchk = nchk
		}
		chk.dchnext = nil
		for vix := 0; vix < dump_chunk_len; vix++ {
			curpob := chk.dchobjects[vix]
			chk.dchobjects[vix] = nil
			if curpob != nil {
				curpob.DumpScanInsideObject(du)
			}
		}
	}
}

func (du *DumperMo) emitDumpedObject(pob *ObjectMo) {
	///@@@ incomplete
}

func (du *DumperMo) LoopDumpEmit() {
	if du == nil || du.dumode != dumod_Scan {
		panic("LoopDumpEmit on non-scanning dumper")
	}
	du.dumode = dumod_Emit
	dso := du.dusetobjects
	if dso == nil {
		panic("LoopDumpEmit: nil dusetobjects")
	}
	for pob, _ := range *dso {
		du.emitDumpedObject(pob)
	}
}
