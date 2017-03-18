// file payloadmo/symbolpayl.go

package payloadmo

import (
	"fmt"
	"log"
	// "serialmo"
	. "objvalmo"
)

type UselessPy struct {
} // end UselessPy

func (sy *UselessPy) DestroyPayl(pob *ObjectMo) {
} // end useless's DestroyPayl

func (sy *UselessPy) DumpScanPayl(pob *ObjectMo, du *DumperMo) {
} // end useless's DumpScanPayl

func (sy *UselessPy) DumpEmitPayl(pob *ObjectMo, du *DumperMo) (pykind string, json interface{}) {
	return "useless", nil
} // end useless's DumpEmitPayl

func (sy *UselessPy) GetPayl(pob *ObjectMo, attrpob *ObjectMo) ValueMo {
	return nil
} // end useless's GetPayl

func (sy *UselessPy) PutPayl(pob *ObjectMo, attrpob *ObjectMo, val ValueMo) error {
	return fmt.Errorf("useless PutPayl pob=%v attrpob=%v val=%v", pob, attrpob, val)
} // end useless's PutPayl

func (sy *UselessPy) DoPayl(pob *ObjectMo, selpob *ObjectMo, args ...ValueMo) error {
	return fmt.Errorf("useless DoPayl pob=%v selpob=%v args=%v", pob, selpob, args)
} // end useless's DoPayl

func loadUseless(kind string, pob *ObjectMo, ld *LoaderMo, jcont interface{}) PayloadMo {
	log.Printf("loadUseless kind=%v pob=%v, cont:%v\n", kind, pob, jcont)
	panic(fmt.Errorf("loadUseless kind=%v pob=%v, cont:%v\n", kind, pob, jcont))
}

func initUseless() {
	log.Printf("initUseless")
	RegisterPayload("useless", PayloadLoaderMo(loadUseless))
} // end initUseless
