// file payloadmo/symbolpayl.go

package payloadmo

import (
	"fmt"
	"log"
	// "serialmo"
	. "objvalmo"
)

type UselessPy struct {
}

func (sy *UselessPy) DestroyPayl(pob *ObjectMo) {
} // end useless's DestroyPayl

func (sy *UselessPy) DumpScanPayl(pob *ObjectMo, du *DumperMo) {
} // end useless's DumpScanPayl

func (sy *UselessPy) DumpEmitPayl(pob *ObjectMo, du *DumperMo) (pykind string, json interface{}) {
	return "useless", nil
} // end useless's DumpEmitPayl

func (sy *UselessPy) LoadPayl(pob *ObjectMo, ld *LoaderMo, paylcont string) {
} // end useless's LoadPayl

func (sy *UselessPy) GetPayl(pob *ObjectMo, attrpob *ObjectMo) ValueMo {
	return nil
} // end useless's GetPayl

func (sy *UselessPy) PutPayl(pob *ObjectMo, attrpob *ObjectMo, val ValueMo) error {
	return fmt.Errorf("useless PutPayl pob=%v attrpob=%v val=%v", pob, attrpob, val)
} // end useless's PutPayl

func (sy *UselessPy) DoPayl(pob *ObjectMo, selpob *ObjectMo, args ...ValueMo) error {
	return fmt.Errorf("useless DoPayl pob=%v selpob=%v args=%v", pob, selpob, args)
} // end useless's DoPayl

func initUseless() {
	log.Printf("initUseless")
} // end initUseless
