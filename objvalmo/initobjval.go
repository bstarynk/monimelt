// file objvalmo/initobjval.go

package objvalmo  // import "github.com/bstarynk/monimelt/objvalmo"

import (
	"log"
)

func init() {
	log.Printf("initobjval start\n")
	initPersist()
	initGlobals()
	log.Printf("initobjval end\n\n")
}
