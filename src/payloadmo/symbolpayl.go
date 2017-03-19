// file payloadmo/symbolpayl.go

package payloadmo

import (
	. "objvalmo"
	//"bytes"
	"fmt"
	llrb "github.com/petar/GoLLRB/llrb"
	"log"
	"regexp"
	"serialmo"
	"sync"
)

type SymbolPy struct {
	syname  string
	syowner *ObjectMo
	syproxy *ObjectMo
	sydata  ValueMo
}

func lessSymbol(a, b interface{}) bool {
	var sya = a.(SymbolPy)
	var syb = b.(SymbolPy)
	return sya.syname < syb.syname
}

func (syl SymbolPy) Less(syr llrb.Item) bool {
	return syl.syname < syr.(SymbolPy).syname
}

const symb_regexp_str = `^[a-zA-Z_][a-zA-Z0-9_]*$`

var symb_regexp *regexp.Regexp = regexp.MustCompile(symb_regexp_str)
var symb_mtx sync.Mutex
var symb_dict llrb.LLRB
var symb_map map[serialmo.IdentMo]*SymbolPy

type jsonSymbol struct {
	Jsyname  string      `json:"syname"`
	Jsyproxy string      `json:"syproxy"`
	Jsydata  interface{} `json:"sydata"`
} // end jsonSymbol

func GetObjectSymbolNamed(nam string) *ObjectMo {
	symb_mtx.Lock()
	defer symb_mtx.Unlock()
	pseudosy := SymbolPy{syname: nam, syowner: nil, syproxy: nil, sydata: nil}
	itsy := symb_dict.Get(pseudosy)
	if itsy == nil {
		return nil
	}
	return itsy.(SymbolPy).syowner
} // end GetObjectSymbolNamed

func GetSymbolNamed(nam string) *SymbolPy {
	symb_mtx.Lock()
	defer symb_mtx.Unlock()
	pseudosy := SymbolPy{syname: nam, syowner: nil, syproxy: nil, sydata: nil}
	itsy := symb_dict.Get(pseudosy)
	if itsy == nil {
		return nil
	}
	return (itsy.(*SymbolPy))
} // end GetSymbolNamed

func HasSymbolNamed(nam string) bool {
	symb_mtx.Lock()
	defer symb_mtx.Unlock()
	pseudosy := SymbolPy{syname: nam, syowner: nil, syproxy: nil, sydata: nil}
	return symb_dict.Has(pseudosy)
} // end HasSymbolNamed

/// return the new added symbol
func AddNewSymbol(nam string, pob *ObjectMo) *SymbolPy {
	var newsy *SymbolPy
	log.Printf("AddNewSymbol nam=%q pob=%v start\n", nam, pob)
	defer log.Printf("AddNewSymbol pob=%v newsy=%v end\n", pob, newsy)
	if pob == nil || nam == "" || !symb_regexp.MatchString(nam) {
		return nil
	}
	symb_mtx.Lock()
	defer symb_mtx.Unlock()
	sy := &SymbolPy{syname: nam, syowner: pob, syproxy: nil, sydata: nil}
	itsy := symb_dict.Get(sy)
	if itsy != nil {
		return nil
	}
	log.Printf("AddNewSymbol pob=%v sy=%#v\n", pob, sy)
	replsy := symb_dict.ReplaceOrInsert(sy)
	log.Printf("AddNewSymbol pob=%v replsy=%#v (%T)\n", pob, replsy, replsy)
	newsy = replsy.(*SymbolPy)
	symb_map[pob.ObId()] = newsy
	return newsy
} // end AddNewSymbol

func (sy *SymbolPy) DestroyPayl(pob *ObjectMo) {
	symb_mtx.Lock()
	defer symb_mtx.Unlock()
	symb_dict.Delete(sy)
	delete(symb_map, pob.ObId())
	sy.syname = ""
	sy.syproxy = nil
	sy.sydata = nil
} // end symbol's DestroyPayl

func (sy *SymbolPy) DumpScanPayl(pob *ObjectMo, du *DumperMo) {
	if sy == nil {
		panic(fmt.Errorf("DumpScanPayl pob=%v nil sy", pob))
	}
	if sy.syproxy != nil {
		du.AddDumpedObject(sy.syproxy)
	}
	if sy.sydata != nil {
		sy.sydata.DumpScan(du)
	}
} // end symbol's DumpScanPayl

func (sy *SymbolPy) DumpEmitPayl(pob *ObjectMo, du *DumperMo) (pykind string, pjson interface{}) {
	var jsy jsonSymbol
	jsy.Jsyname = sy.syname
	if sy.syproxy != nil && du.EmitObjptr(sy.syproxy) {
		jsy.Jsyproxy = sy.syproxy.ToString()
	}
	jsy.Jsydata = ValToJson(du, sy.sydata)
	return "symbol", jsy
} // end symbol's DumpEmitPayl

func loadSymbol(kind string, pob *ObjectMo, ld *LoaderMo, jcont interface{}) PayloadMo {
	log.Printf("loadSymbol kind=%v pob=%v, jcont:%v\n", kind, pob, jcont)
	var syname string
	var syproxidstr string
	var syproxpob *ObjectMo
	var sydata ValueMo
	var jsyname interface{}
	var jsyproxy interface{}
	var jsydata interface{}
	var jcontmap map[string]interface{}
	var ok bool
	var err error
	jcontmap, ok = jcont.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("loadSymbol pob=%v bad jcont=%v", pob, jcont))
	}
	jsyname, ok = jcontmap["syname"]
	if !ok {
		panic(fmt.Errorf("loadSymbol pob=%v missing syname in jcontmap=%v", pob, jcontmap))
	}
	syname, ok = jsyname.(string)
	if !ok {
		panic(fmt.Errorf("loadSymbol pob=%v bad syname in jcontmap=%v", pob, jcontmap))
	}
	jsyproxy, ok = jcontmap["syproxy"]
	if ok {
		syproxidstr, ok = jsyproxy.(string)
		if !ok {
			panic(fmt.Errorf("loadSymbol pob=%v bad syproxy in jcontmap=%v", pob, jcontmap))
		}
		syproxpob, err = ld.ParseObjptr(syproxidstr)
		if err != nil {
			panic(fmt.Errorf("loadSymbol pob=%v wrong syproxy in jcontmap=%v", pob, jcontmap))
		}
	}
	jsydata, ok = jcontmap["sydata"]
	if ok {
		sydata, err = JasonParseVal(ld, jsydata)
		if err != nil {
			panic(fmt.Errorf("loadSymbol pob=%v bad sydata in jcontmap=%v : %v", pob, jcontmap, err))
		}
	}
	var sy *SymbolPy
	sy = AddNewSymbol(syname, pob)
	sy.sydata = sydata
	sy.syproxy = syproxpob
	log.Printf("loadSymbol pob=%v sy=%#v\n", pob, sy)
	return sy
} // end loadSymbol

func (sy *SymbolPy) GetPayl(pob *ObjectMo, attrpob *ObjectMo) ValueMo {
	panic("symbol's GetPayl unimplemented")
} // end symbol's GetPayl

func (sy *SymbolPy) PutPayl(pob *ObjectMo, attrpob *ObjectMo, val ValueMo) error {
	panic("symbol's PutPayl unimplemented")
} // end symbol's PutPayl

func (sy *SymbolPy) DoPayl(pob *ObjectMo, selpob *ObjectMo, args ...ValueMo) error {
	panic("symbol's DoPayl unimplemented")
} // end symbol's DoPayl

func initSymbol() {
	log.Printf("initSymbol")
	RegisterPayload("symbol", PayloadLoaderMo(loadSymbol))
} // end initSymbol
