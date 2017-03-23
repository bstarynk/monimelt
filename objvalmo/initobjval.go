// file objvalmo/initobjval.go

package objvalmo

import (
	"log"
)

func init() {
	log.Printf("initobjval start\n")
	initPersist()
	initGlobals()
	log.Printf("initobjval end\n\n")
}
