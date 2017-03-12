// file objvalmo/persist.go

package objvalmo

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	gosqlite "github.com/mattn/go-sqlite3"
	"log"
	"os"
	osexec "os/exec"
	"regexp"
	"runtime"
	"serialmo"
	"strings"
	"time"
)

const DefaultGlobalDbname = "monimelt_state"
const DefaultUserDbname = "monimelt_user"

const SqliteProgram = "sqlite3"
const GlobalObjects = true
const UserObjects = false

func initPersist() {
	EnableSqliteLog()
	libv, libvnum, srcid := gosqlite.Version()
	log.Printf("persist-init libv=%s libvnum=%d srcid=%s\n",
		libv, libvnum, srcid)
}

type LoaderMo struct {
	ldglobaldb *sql.DB
	lduserdb   *sql.DB
	ldobjmap   map[serialmo.IdentMo]*ObjectMo
}

var validpath_regexp *regexp.Regexp

const validpath_regexp_str = `^[a-zA-Z0-9_/.+-]*$`

func validpath(path string) bool {
	if validpath_regexp == nil {
		validpath_regexp = regexp.MustCompile(validpath_regexp_str)
	}
	// check that path has no double dots like ..
	return validpath_regexp.MatchString(path) && !strings.Contains(path, "..")
}

func OpenLoaderFromFiles(globalpath string, userpath string) *LoaderMo {
	{
		var stabuf [2048]byte
		stalen := runtime.Stack(stabuf[:], true)
		log.Printf("OpenLoaderFromFiles start globalpath=%s userpath=%s\n...stack:\n%s\n\n\n",
			globalpath, userpath, string(stabuf[:stalen]))
	}
	if !validpath(globalpath) {
		panic(fmt.Errorf("OpenLoaderFromFiles invalid global path %s",
			globalpath))
	}
	if _, err := os.Stat(globalpath); err != nil {
		panic(fmt.Errorf("OpenLoaderFromFiles wrong global path %s - %v",
			globalpath, err))
	}
	if len(userpath) > 0 {
		if !validpath(userpath) {
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
	l.ldobjmap = make(map[serialmo.IdentMo]*ObjectMo)
	return l
} /// end OpenLoaderFromFiles

func (l *LoaderMo) ParseObjptr(oidstr string) (*ObjectMo, error) {
	var pob *ObjectMo
	log.Printf("loader ParseObjptr oidstr=%s\n", oidstr)
	defer log.Printf("loader ParseObjptr oidstr=%s pob=%v", oidstr, pob)
	oid, err := serialmo.IdFromString(oidstr)
	if err != nil {
		return nil, err
	}
	pob = l.ldobjmap[oid]
	if pob == nil {
		return nil, fmt.Errorf("loader ParseObjptr not found %q", oidstr)
	}
	return pob, nil
} // end loader ParseObjptr

func (l *LoaderMo) create_objects(globflag bool) {
	log.Printf("create_objects start globflag=%t\n", globflag)
	defer log.Printf("create_objects end globflag=%t\n\n", globflag)
	var qr *sql.Rows
	var err error
	const sql_selcreated = "SELECT ob_id FROM t_objects"
	if globflag {
		qr, err = l.ldglobaldb.Query(sql_selcreated)
	} else {
		qr, err = l.lduserdb.Query(sql_selcreated)
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
		l.ldobjmap[oid] = pob
		log.Printf("create_objects pob=%v\n", pob)
	}
	err = qr.Err()
	if err != nil {
		panic(fmt.Errorf("persistmo.create_objects final %v", err))
	}
} // end create_objects

func (l *LoaderMo) fill_content_objects(globflag bool) {
	log.Printf("fill_content_objects start globflag=%t\n", globflag)
	defer log.Printf("fill_content_objects end globflag=%t\n", globflag)
	var qr *sql.Rows
	var err error
	const sql_selfillcontent = `SELECT ob_id, ob_mtime, ob_jsoncont FROM t_objects`
	if globflag {
		qr, err = l.ldglobaldb.Query(sql_selfillcontent)
	} else {
		qr, err = l.lduserdb.Query(sql_selfillcontent)
	}
	if err != nil {
		panic(fmt.Errorf("loader: fill_content_objects failure %v", err))
	}
	defer qr.Close()
	for qr.Next() {
		var idstr string
		var mtim int64
		var jcontstr string
		err = qr.Scan(&idstr, &mtim, &jcontstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.fill_content_objects failure %v", err))
		}
		oid, err := serialmo.IdFromString(idstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.fill_content_objects bad id %s: %v", idstr, err))
		}
		pob := l.ldobjmap[oid]
		if pob == nil {
			panic(fmt.Errorf("persistmo.fill_content_objects unknown id %s: %v", idstr, err))
		}
		pob.UnsyncPutMtime(mtim)
		var jcont jsonObContent
		if err := json.Unmarshal(([]byte)(jcontstr), &jcont); err != nil {
			panic(fmt.Errorf("persistmo.fill_content_objects bad content for id %s: %v", idstr, err))
		}
		nbat := len(jcont.Jattrs)
		for atix := 0; atix < nbat; atix++ {
			curatid := jcont.Jattrs[atix].Jat
			curjval := jcont.Jattrs[atix].Jva
			log.Printf("atix=%d curatid=%v curjval=%v\n", atix, curatid, curjval)
		}
		// do something with jcont
		log.Printf("@@@fill_content_objects pob=%v mtim=%v jcont=%#v %T\n\n", pob, mtim, jcont, jcont)
		for pix, jcurpair := range jcont.Jattrs {
			log.Printf("fill_content_objects pob=%v pix=%d jcurpair=%v %T\n",
				pob, pix, jcurpair, jcurpair)
			pobat, err := l.ParseObjptr(jcurpair.Jat)
			log.Printf("fill_content_objects pob=%v pix#%d pobat=%v/%T, err=%v\n",
				pob, pix, pobat, pobat, err)
			/**@@@ FIXME
						atval, err := JasonParseValue(l,jcurpair.Jva)
						log.Printf("fill_content_objects pob=%v pix#%d atval=%v/%T, err=%v\n",
							pob, pix, atval, err)
			                        ***/

		}
		for cix, jcurcomp := range jcont.Jcomps {
			log.Printf("fill_content_objects pob=%v cix=%d jcurcomp=%v %T\n",
				pob, cix, jcurcomp, jcurcomp)
		}
	}
	if err = qr.Err(); err != nil {
		panic(fmt.Errorf("persistmo.fill_content_objects err %v", err))
	}
} // end fill_content_objects

func (l *LoaderMo) fill_payload_objects(globflag bool) {
	log.Printf("fill_payload_objects start globflag=%t\n", globflag)
	defer log.Printf("fill_payload_objects end globflag=%t\n", globflag)
	var qr *sql.Rows
	var err error
	const sql_selfillcontent = `SELECT ob_id, ob_paylkind, ob_paylcont 
FROM t_objects WHERE ob_paylkind != ""`
	if globflag {
		qr, err = l.ldglobaldb.Query(sql_selfillcontent)
	} else {
		qr, err = l.lduserdb.Query(sql_selfillcontent)
	}
	if err != nil {
		panic(fmt.Errorf("loader: fill_payload_objects failure %v", err))
	}
	defer qr.Close()
	for qr.Next() {
		var idstr string
		var paylkind string
		var jpaylstr string
		err = qr.Scan(&idstr, &paylkind, &jpaylstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.fill_payload_objects failure %v", err))
		}
		oid, err := serialmo.IdFromString(idstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.fill_payload_objects bad id %s: %v", idstr, err))
		}
		pob := l.ldobjmap[oid]
		if pob == nil {
			panic(fmt.Errorf("persistmo.fill_payload_objects unknown id %s: %v", idstr, err))
		}
		log.Printf("fill_payload_objects @@incomplete pob=%v paylkind=%s\n", pob, paylkind)
	}
	log.Printf("fill_payload_objects end globflag=%t\n", globflag)
} // end fill_payload_objects

func (l *LoaderMo) bind_globals(globflag bool) {
	log.Printf("bind_globals start globflag=%t\n", globflag)
	defer log.Printf("bind_globals end globflag=%t\n\n", globflag)
	var qr *sql.Rows
	var err error
	const sql_selglobals = `SELECT glob_name, glob_oid FROM t_globals WHERE glob_oid!=""`
	if globflag {
		qr, err = l.ldglobaldb.Query(sql_selglobals)
	} else {
		qr, err = l.lduserdb.Query(sql_selglobals)
	}
	if err != nil {
		panic(fmt.Errorf("loader: bind_globals failure %v", err))
	}
	defer qr.Close()
	for qr.Next() {
		var globname string
		var globidstr string
		err = qr.Scan(&globname, &globidstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.bind_globals failure %v", err))
		}
		gloid, err := serialmo.IdFromString(globidstr)
		if err != nil {
			panic(fmt.Errorf("persistmo.bind_globals bad id %s: %v", globidstr, err))
		}
		glpob := l.ldobjmap[gloid]
		if glpob == nil {
			panic(fmt.Errorf("persistmo.bind_globals unknown id %s: %v", globidstr, err))
		}
		pglovar := GlobalVariableAddress(globname)
		if pglovar == nil {
			panic(fmt.Errorf("persistmo.bind_globals unknown global %s", globname))
		}
		*pglovar = glpob
	}
} // end bind_globals

func (ld *LoaderMo) Load() {
	{
		var stabuf [1024]byte
		stalen := runtime.Stack(stabuf[:], false)
		log.Printf("loader Load start ld=%#v\n...stack:\n%s\n\n\n",
			ld, string(stabuf[:stalen]))
	}
	defer log.Printf("Load end ld=%v\n", ld)
	if ld == nil {
		return
	}
	ld.create_objects(GlobalObjects)
	if ld.lduserdb != nil {
		ld.create_objects(UserObjects)
	}
	log.Printf("Load after create_objects ld=%v\n", ld)
	ld.fill_content_objects(GlobalObjects)
	if ld.lduserdb != nil {
		ld.fill_content_objects(UserObjects)
	}
	log.Printf("Load after fill_content_objects ld=%v\n", ld)
	ld.fill_payload_objects(GlobalObjects)
	if ld.lduserdb != nil {
		ld.fill_payload_objects(UserObjects)
	}
	log.Printf("Load after fill_payload_objects ld=%v\n", ld)
	ld.bind_globals(GlobalObjects)
	if ld.lduserdb != nil {
		ld.bind_globals(UserObjects)
	}
	log.Printf("Load after bind_globals ld=%v\n", ld)
} // end Load

func (ld *LoaderMo) Close() {
	{
		var stabuf [1024]byte
		stalen := runtime.Stack(stabuf[:], false)
		log.Printf("loader Close start ld=%#v\n...stack:\n%s\n\n\n",
			ld, string(stabuf[:stalen]))
	}
	if ld == nil {
		return
	}
	if ud := ld.lduserdb; ud != nil {
		ld.lduserdb = nil
		ud.Close()
	}
	if gd := ld.ldglobaldb; gd != nil {
		ld.ldglobaldb = nil
		gd.Close()
	}
	/// clear the object map
	ld.ldobjmap = nil
} // end Close

func LoadFromDirectory(dirname string) {
	defer log.Printf("LoadFromDirectory %s end\n\n", dirname)
	{
		var stabuf [2048]byte
		stalen := runtime.Stack(stabuf[:], true)
		log.Printf("LoadFromDirectory dirname=%s\n...stack:\n%s\n\n\n",
			dirname, string(stabuf[:stalen]))
	}
	if dirname == "" {
		dirname = "."
	}
	dl := len(dirname)
	if dirname[dl-1] != '/' {
		dirname = dirname + "/"
	}
	if dinf, err := os.Stat(dirname); err != nil || !dinf.Mode().IsDir() {
		panic(fmt.Errorf("LoadFromDirectory bad dirname %s - %v, %v", dirname, err, dinf))
	}
	glodbpath := dirname + DefaultGlobalDbname + ".sqlite"
	glodbinf, err := os.Stat(glodbpath)
	if err != nil || !glodbinf.Mode().IsRegular() {
		panic(fmt.Errorf("LoadFromDirectory bad global db %s - %v, %v", glodbpath, err, glodbinf))
	}
	glosqlpath := dirname + DefaultGlobalDbname + ".sql"
	glosqlinf, err := os.Stat(glosqlpath)
	if err == nil {
		if glosqlinf.ModTime().Before(glodbinf.ModTime()) {
			panic(fmt.Errorf("LoadFromDirectory global sql file %s [%v] older than db file %s [%v]",
				glosqlpath, glosqlinf.ModTime(), glodbpath, glodbinf.ModTime()))
		}
	}
	usrdbpath := dirname + DefaultUserDbname + ".sqlite"
	usrsqlpath := dirname + DefaultUserDbname + ".sql"
	if usrdbinf, err := os.Stat(usrdbpath); err != nil || !usrdbinf.Mode().IsRegular() {
		log.Printf("LoadFromDirectory missing or bad user db %s\n", usrdbpath)
		usrdbpath = ""
		usrsqlpath = ""
	} else {
		if usrsqlinf, err := os.Stat(usrsqlpath); err == nil {
			if usrsqlinf.ModTime().Before(usrdbinf.ModTime()) {
				panic(fmt.Errorf("LoadFromDirectory user sql file %s [%v] older than db file %s [%v]",
					usrsqlpath, glosqlinf.ModTime(), usrdbpath, usrdbinf.ModTime()))
			}
		}
	}
	ld := OpenLoaderFromFiles(glodbpath, usrdbpath)
	defer ld.Close()
	ld.Load()
	log.Printf("done LoadFromDirectory %s\n", dirname)
} // end LoadFromDirectory

////////////////////////////////////////////////////////////////
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
	dutime       time.Time
	dumode       uint
	dudirname    string
	dutempsuffix string
	duglobaldb   *sql.DB
	duuserdb     *sql.DB
	dustobuser   *sql.Stmt
	dustobglob   *sql.Stmt
	dufirstchk   *dumpChunk
	dulastchk    *dumpChunk
	dusetobjects map[*ObjectMo]uint8
}

const sql_create_t_params = `CREATE TABLE IF NOT EXISTS t_params 
 (par_name VARCHAR(35) PRIMARY KEY ASC NOT NULL UNIQUE, 
  par_value TEXT NOT NULL);`

const sql_create_t_objects = `CREATE TABLE IF NOT EXISTS t_objects
 (ob_id VARCHAR(26) PRIMARY KEY ASC NOT NULL UNIQUE,
  ob_mtime INT NOT NULL,
  ob_jsoncont TEXT NOT NULL,
  ob_paylkind VARCHAR(40) NOT NULL,
  ob_paylcont TEXT NOT NULL);`

const sql_create_t_globals = `CREATE TABLE IF NOT EXISTS t_globals
 (glob_name VARCHAR(80) PRIMARY KEY ASC NOT NULL UNIQUE,
  glob_oid VARCHAR(26)  NOT NULL);`

const sql_insert_t_objects = `INSERT INTO t_objects VALUES (?, ?, ?, ?, ?);`

const sql_insert_t_globals = `INSERT INTO t_globals VALUES (?, ?)`

func (du DumperMo) create_tables(globflag bool) {
	var db *sql.DB
	log.Printf("create_table globflag=%v dir=%s\n", globflag, du.dudirname)
	if globflag {
		db = du.duglobaldb
	} else {
		db = du.duuserdb
	}
	if db == nil {
		panic(fmt.Errorf("create_tables no db in directory %s", du.dudirname))
	}
	log.Printf("create_table db=%v sql_create_t_params=%q\n", db, sql_create_t_params)
	rcparam, err := db.Exec(sql_create_t_params)
	log.Printf("create_table rcparam=%v err=%v\n", rcparam, err)
	if err != nil {
		panic(fmt.Errorf("create_tables failure in directory %s for t_params creation %v",
			du.dudirname, err))
	}
	_, err = db.Exec(sql_create_t_objects)
	if err != nil {
		panic(fmt.Errorf("create_tables failure in directory %s for t_objects creation %v",
			du.dudirname, err))
	}
	_, err = db.Exec(sql_create_t_globals)
	if err != nil {
		panic(fmt.Errorf("create_tables failure in directory %s for t_globals creation %v",
			du.dudirname, err))
	}
} // end create_tables

func (du *DumperMo) AddDumpedObject(pob *ObjectMo) {
	log.Printf("AddDumpedObject start pob=%v\n", pob)
	defer log.Printf("AddDumpedObject end pob=%v\n", pob)
	if du.dumode != dumod_Scan {
		panic("AddDumpedObject in non-scanning dumper")
	}
	if pob == nil {
		return
	}
	spo := pob.UnsyncSpaceNum() /// should be unsync
	if spo == SpaTransient {
		return
	}
	log.Printf("AddDumpedObject for pob=%v du.dusetobjects=%v\n", pob, du.dusetobjects)
	if _, found := du.dusetobjects[pob]; found {
		log.Printf("AddDumpedObject found pob=%v\n", pob)
		return
	}
	log.Printf("AddDumpedObject pob=%v before dusetobjects=%p=%#v\n", pob, du.dusetobjects, du.dusetobjects)
	du.dusetobjects[pob] = spo
	log.Printf("AddDumpedObject pob=%v after dusetobjects=%#v\n", pob, du.dusetobjects)
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
	log.Printf("AddDumpedObject pob=%v lchk=%#v putix=%d\n", pob, lchk, putix)
	if putix >= 0 {
		lchk.dchobjects[putix] = pob
		return
	}
	nchk := new(dumpChunk)
	lchk.dchnext = nchk
	du.dulastchk = nchk
	nchk.dchobjects[0] = pob
	log.Printf("AddDumpedObject pob=%v nchk=%#v", pob, nchk)
} // end AddDumpedObject

func OpenDumperDirectory(dirpath string) *DumperMo {
	if !validpath(dirpath) {
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
	dtempsuf := fmt.Sprintf("+%s_p%d.tmp", serialmo.RandomSerial().ToString(), os.Getpid())
	log.Printf("OpenDumperDirectory dirpath=%s dtempsuf=%s\n", dirpath, dtempsuf)
	globtemppath := fmt.Sprintf("%s/%s.sqlite%s", dirpath, DefaultGlobalDbname, dtempsuf)
	usertemppath := fmt.Sprintf("%s/%s.sqlite%s", dirpath, DefaultUserDbname, dtempsuf)
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
	du.dutime = time.Now()
	du.dudirname = dirpath
	du.dutempsuffix = dtempsuf
	du.duglobaldb = glodb
	du.duuserdb = usrdb
	du.dusetobjects = make(map[*ObjectMo]uint8)
	du.create_tables(GlobalObjects)
	du.create_tables(UserObjects)
	du.dustobglob, err = glodb.Prepare(sql_insert_t_objects)
	if err != nil {
		// this should never happen
		panic(fmt.Errorf("OpenDumperDirectory failed to prepare global %s t_object insertion - %v", globtemppath, err))
	}
	du.dustobuser, err = usrdb.Prepare(sql_insert_t_objects)
	if err != nil {
		// this should never happen
		panic(fmt.Errorf("OpenDumperDirectory failed to prepare global %s t_object insertion - %v", usertemppath, err))
	}
	log.Printf("OpenDumperDirectory result du=%#v\n", du)
	return du
} // end OpenDumperDirectory

func (du *DumperMo) StartDumpScan() {
	log.Printf("StartDumpScan begin du=%#v\n", du)
	defer log.Printf("StartDumpScan end du=%#v\n", du)
	if du == nil || du.dumode != dumod_Idle {
		panic("StartDumpScan on non-idle dumper")
	}
	du.dumode = dumod_Scan
	DumpScanPredefined(du)
	log.Printf("StartDumpScan after scan-predefined du=%v\n", du)
	DumpScanGlobalVariables(du)
} // end StartDumpScan

func (du *DumperMo) IsDumpedObject(pob *ObjectMo) bool {
	_, found := du.dusetobjects[pob]
	return found
}

func (du *DumperMo) LoopDumpScan() {
	log.Printf("LoopDumpScan begin du=%#v\n", du)
	defer log.Printf("LoopDumpScan end du=%#v\n", du)
	if du == nil || du.dumode != dumod_Scan {
		panic("LoopDumpScan on non-scanning dumper")
	}
	var chk *dumpChunk
	var nchk *dumpChunk
	for chk = du.dufirstchk; chk != nil; chk = nchk {
		log.Printf("LoopDumpScan chk=%#v\n", chk)
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
				log.Printf("LoopDumpScan vix=%d curpob=%v\n", vix, curpob)
				curpob.DumpScanInsideObject(du)
			}
		}
	}
} // end LoopDumpScan

type jsonAttrEntry struct {
	Jat string      `json:"at"`
	Jva interface{} `json:"va"`
}

type jsonObContent struct {
	Jattrs []jsonAttrEntry `json:"attrs"`
	Jcomps []interface{}   `json:"comps"`
}

func (du *DumperMo) emitDumpedObject(pob *ObjectMo, spa uint8) {
	log.Printf("emitDumpedObject start pob=%v spa=%d\n", pob, spa)
	defer log.Printf("emitDumpedObject end pob=%v spa=%d\n\n", pob, spa)
	if du == nil || du.dumode != dumod_Emit {
		panic("emitDumpedObject bad dumper")
	}
	if pob == nil {
		panic("emitDumpedObject nil object")
	}
	if spa == SpaTransient || spa >= Spa_Last {
		panic("emitDumpedObject bad spa")
	}
	pobidstr := pob.ToString()
	pob.obmtx.Lock()
	defer pob.obmtx.Unlock()
	/// dump the attrbitues
	nbat := len(pob.obattrs)
	log.Printf("emitDumpedObject pob=%v nbat=%d\n", pob, nbat)
	var jattrs []jsonAttrEntry
	jattrs = make([]jsonAttrEntry, 0, nbat)
	for atob, atva := range pob.obattrs {
		if atva == nil {
			continue
		}
		if !du.EmitObjptr(atob) {
			continue
		}
		jpair := jsonAttrEntry{Jat: atob.ToString(), Jva: ValToJson(du, atva)}
		log.Printf("emitDumpedObject pob=%v atob=%v atva=%v jpair=%v\n",
			pob, atob, atva, jpair)
		jattrs = append(jattrs, jpair)
	}
	/// dump the components
	nbcomp := len(pob.obcomps)
	log.Printf("emitDumpedObject pob=%v nbcomp=%d\n", pob, nbcomp)
	var jcomps []interface{}
	jcomps = make([]interface{}, 0, nbcomp)
	for cix, cva := range pob.obcomps {
		jva := ValToJson(du, cva)
		log.Printf("emitDumpedObject pob=%v cix=%d cva=%v jva=%v\n",
			pob, cix, cva, jva)
		jcomps = append(jcomps, jva)
	}
	/// construct and encode the content
	jcontent := jsonObContent{Jattrs: jattrs, Jcomps: jcomps}
	log.Printf("emitDumpedObject pob=%v jcontent=%v\n", pob, jcontent)
	var contbuf bytes.Buffer
	contenc := json.NewEncoder(&contbuf)
	contbuf.WriteByte('\n')
	contenc.SetIndent("", " ")
	contenc.Encode(jcontent)
	log.Printf("emitDumpedObject pob=%v contbuf=%s\n", pob, contbuf.String())
	//contbuf.WriteByte('\n')
	/// encode the payload
	var paylkindstr string
	var jpayljson interface{}
	if pob.obpayl != nil {
		paylkindstr, jpayljson = (*pob.obpayl).DumpEmitPayl(pob, du)
	}
	var paylbuf bytes.Buffer
	if len(paylkindstr) > 0 {
		paylenc := json.NewEncoder(&paylbuf)
		paylbuf.WriteByte('\n')
		paylenc.SetIndent("", " ")
		paylenc.Encode(jpayljson)
		//paylbuf.WriteByte('\n')
	}
	/// should now insert in the appropriate database
	var stmt *sql.Stmt
	if spa == SpaUser {
		stmt = du.dustobuser
	} else {
		stmt = du.dustobglob
	}
	var err error
	_, err = stmt.Exec(pobidstr,
		fmt.Sprintf("%d", pob.UnsyncMtime()),
		contbuf.String(),
		paylkindstr,
		paylbuf.String())
	if err != nil {
		panic(fmt.Errorf("emitDumpedObject insertion failed for %s - %v", pobidstr, err))
	}
} // end emitDumpedObject

func (du *DumperMo) EmitObjptr(pob *ObjectMo) bool {
	_, found := du.dusetobjects[pob]
	return found
}

func (du *DumperMo) DumpEmit() {
	log.Printf("DumpEmit start du=%#v\n", du)
	defer log.Printf("DumpEmit end du=%#v\n", du)
	if du == nil || du.dumode != dumod_Scan {
		panic("DumpEmit on non-scanning dumper")
	}
	du.dumode = dumod_Emit
	globstmt, err := du.duglobaldb.Prepare(sql_insert_t_globals)
	if err != nil {
		panic(fmt.Errorf("DumpEmit failed to prepare t_globals insertion %v", err))
	}
	defer globstmt.Close()
	usrglobstmt, err := du.duuserdb.Prepare(sql_insert_t_globals)
	if err != nil {
		panic(fmt.Errorf("DumpEmit failed to prepare user t_globals insertion %v", err))
	}
	defer usrglobstmt.Close()
	// emit all objects
	dso := du.dusetobjects
	if dso == nil {
		panic("DumpEmit: nil dusetobjects")
	}
	for pob, sp := range dso {
		du.emitDumpedObject(pob, sp)
	}
	/// emit the global variables
	globnames := NamesGlobalVariables()
	for _, gname := range globnames {
		gad := GlobalVariableAddress(gname)
		if gad == nil {
			continue
		}
		gpob := *gad
		if gpob == nil {
			continue
		}
		gsp := du.dusetobjects[gpob]
		if gsp == SpaGlobal || gsp == SpaPredefined {
			_, err := globstmt.Exec(gname, gpob.ToString())
			if err != nil {
				panic(fmt.Errorf("DumpEmit failed to insert global %s - %v", gname, err))
			}
		} else if gsp == SpaUser {
			_, err := usrglobstmt.Exec(gname, gpob.ToString())
			if err != nil {
				panic(fmt.Errorf("DumpEmit failed to insert global %s - %v", gname, err))
			}
		}
	}
} // end DumpEmit

func (du *DumperMo) renameWithBackup(fpath string) {
	log.Printf("renameWithBackup fpath=%s tempsuffix=%s\n", fpath, du.dutempsuffix)
	if du == nil {
		panic(fmt.Errorf("renameWithBackup nil du fpath=%q\n", fpath))
	}
	tmpath := du.dudirname + "/" + fpath + du.dutempsuffix
	newpath := du.dudirname + "/" + fpath
	backupath := newpath + "~"
	if _, err := os.Stat(backupath); err == nil {
		os.Rename(backupath, backupath+"~")
	}
	if _, err := os.Stat(newpath); err == nil {
		os.Rename(newpath, backupath)
	}
	if err := os.Rename(tmpath, newpath); err != nil {
		panic(fmt.Errorf("renameWithBackup dumpdir %s failed for %s  -> %s - %v", du.dudirname, tmpath, newpath, err))
	}
} // end renameWithBackup

func (du *DumperMo) Close() {
	{
		var stabuf [2048]byte
		stalen := runtime.Stack(stabuf[:], true)
		log.Printf("dumper Close start du=%#v stack:\n%s\n\n\n",
			du, string(stabuf[:stalen]))
	}
	defer log.Printf("dumper Close end\n")
	if du == nil {
		return
	}
	var nbob int
	if du.dusetobjects != nil {
		nbob = len(du.dusetobjects)
	}
	du.dusetobjects = nil
	du.dulastchk = nil
	du.dufirstchk = nil
	du.dustobglob.Close()
	du.dustobglob = nil
	if du.dustobuser != nil {
		du.dustobuser.Close()
		du.dustobuser = nil
	}
	if du.duuserdb != nil {
		du.duuserdb.Close()
		du.duuserdb = nil
	}
	if du.duglobaldb != nil {
		du.duglobaldb.Close()
		du.duglobaldb = nil
	}
	var err error
	globtempdb := fmt.Sprintf("%s/%s.sqlite%s", du.dudirname,
		DefaultGlobalDbname, du.dutempsuffix)
	globtempsql := fmt.Sprintf("%s/%s.sql%s", du.dudirname,
		DefaultGlobalDbname, du.dutempsuffix)
	globoutput := fmt.Sprintf(".output %s", globtempsql)
	globstacmt := fmt.Sprintf("-- generated monimelt global dumpfile %s.sql", DefaultGlobalDbname)
	globstaprint := fmt.Sprintf(".print %q", globstacmt)
	globendcmt := fmt.Sprintf("-- end of monimelt global dumpfile %s.sql", DefaultGlobalDbname)
	globendprint := fmt.Sprintf(".print %q", globendcmt)
	var cmd *osexec.Cmd
	cmd = osexec.Command(SqliteProgram, globtempdb, globoutput, globstaprint, ".dump", globendprint)
	if err = cmd.Run(); err != nil {
		panic(fmt.Errorf("dumper Close failed to run global dump %s - %v",
			cmd, err))
	}
	cmd = nil
	usertempdb := fmt.Sprintf("%s/%s.sqlite%s", du.dudirname,
		DefaultUserDbname, du.dutempsuffix)
	usertempsql := fmt.Sprintf("%s/%s.sql%s", du.dudirname,
		DefaultUserDbname, du.dutempsuffix)
	useroutput := fmt.Sprintf(".output %s", usertempsql)
	userstacmt := fmt.Sprintf("-- generated monimelt user dumpfile %s.sql", DefaultUserDbname)
	userstaprint := fmt.Sprintf(".print %q", userstacmt)
	userendcmt := fmt.Sprintf("-- end of monimelt user dumpfile %s.sql", DefaultUserDbname)
	userendprint := fmt.Sprintf(".print %q", userendcmt)
	cmd = osexec.Command(SqliteProgram, usertempdb, useroutput, userstaprint, ".dump", userendprint)
	if err = cmd.Run(); err != nil {
		panic(fmt.Errorf("dumper Close failed to run user dump %s - %v",
			cmd, err))
	}
	cmd = nil
	nowt := du.dutime
	os.Chtimes(globtempdb, nowt, nowt)
	os.Chtimes(globtempsql, nowt, nowt)
	os.Chtimes(usertempdb, nowt, nowt)
	os.Chtimes(usertempsql, nowt, nowt)
	du.renameWithBackup(DefaultGlobalDbname + ".sql")
	du.renameWithBackup(DefaultUserDbname + ".sql")
	du.renameWithBackup(DefaultGlobalDbname + ".sqlite")
	du.renameWithBackup(DefaultUserDbname + ".sqlite")
	log.Printf("done dump of %d objects in %s\n", nbob, du.dudirname)
} // end dumper Close

func DumpIntoDirectory(dirname string) {
	log.Printf("DumpIntoDirectory start dirname=%s\n\n", dirname)
	defer log.Printf("DumpIntoDirectory ended dirname=%s\n\n", dirname)
	var du *DumperMo
	du = OpenDumperDirectory(dirname)
	defer du.Close()
	log.Printf("==== DumpIntoDirectory before StartDumpScan du=%#v\n\n", du)
	du.StartDumpScan()
	log.Printf("==== DumpIntoDirectory before LoopDumpScan du=%#v\n\n", du)
	du.LoopDumpScan()
	log.Printf("==== DumpIntoDirectory before DumpEmit du=%#v\n\n", du)
	du.DumpEmit()
	log.Printf("==== DumpIntoDirectory final du=%#v\n\n", du)
} // end DumpIntoDirectory
