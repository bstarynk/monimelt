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

func (sy *SymbolPy) LoadPayl(pob *ObjectMo, ld *LoaderMo, paylcont string) {
} // end symbol's LoadPayl

func (sy *SymbolPy) GetPayl(pob *ObjectMo, attrpob *ObjectMo) ValueMo {
	panic("symbol's GetPayl unimplemented")
} // end symbol's GetPayl

func (sy *SymbolPy) PutPayl(pob *ObjectMo, attrpob *ObjectMo, val ValueMo) error {
	panic("symbol's PutPayl unimplemented")
} // end symbol's PutPayl

func (sy *SymbolPy) DoPayl(pob *ObjectMo, selpob *ObjectMo, args ...ValueMo) error {
	panic("symbol's DoPayl unimplemented")
} // end symbol's DoPayl

func buildSymbol(kind string, pob *ObjectMo) *PayloadMo {
	log.Printf("buildSymbol kind=%v pob=%v\n", kind, pob)
	var sy *SymbolPy
	sy = new(SymbolPy)
	log.Printf("buildSymbol pob=%v sy=%#v\n", pob, sy)
	panic("buildSymbol dont know how to return sy")
}

func initSymbol() {
	log.Printf("initSymbol")
	RegisterPayload("symbol", PayloadBuilderMo(buildSymbol))
} // end initSymbol
