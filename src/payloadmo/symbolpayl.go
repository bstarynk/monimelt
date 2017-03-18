// file payloadmo/symbolpayl.go

package payloadmo

import (
	. "objvalmo"
	//"bytes"
	//"fmt"
	"log"
	//"serialmo"
)

type SymbolPy struct {
	syname  string
	syproxy *ObjectMo
	sydata  ValueMo
}

func (sy *SymbolPy) DestroyPayl(pob *ObjectMo) {
} // end symbol's DestroyPayl

func (sy *SymbolPy) DumpScanPayl(pob *ObjectMo, du *DumperMo) {
} // end symbol's DumpScanPayl

func (sy *SymbolPy) DumpEmitPayl(pob *ObjectMo, du *DumperMo) (pykind string, pjson interface{}) {
	panic("symbol's DumpEmitPayl unimplemented")
} // end symbol's DumpEmitPayl

func loadSymbol(kind string, pob *ObjectMo, ld *LoaderMo, jcont interface{}) PayloadMo {
	log.Printf("loadSymbol kind=%v pob=%v, jcont:%v\n", kind, pob, jcont)
	var sy *SymbolPy
	sy = new(SymbolPy)
	log.Printf("loadSymbol pob=%v sy=%#v\n", pob, sy)
	return sy
	// panic("loadSymbol dont know how to return sy")
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
