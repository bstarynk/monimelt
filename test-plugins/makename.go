/// file makename.go, a plugin
package main // import "github.com/bstarynk/monimelt/test-plugins/"

import (
	"log"
	// our packages
	"objvalmo"    // import "github.com/bstarynk/monimelt/objvalmo"
	"payloadmo"   // import "github.com/bstarynk/monimelt/payloadmo"
)

var namob *objvalmo.ObjectMo

func init() {
	namob = objvalmo.Predef_02hL3RuX4x6_6y6PTK9vZs7()
	log.Printf("makename init namob=%v\n", namob)
}

func DoMonimelt() {
	log.Printf("makename DoMonimelt namob=%v\n", namob)
	sy := payloadmo.AddNewSymbol("name", namob)
	log.Printf("makename namob=%v sy=%v\n", namob, sy)
}
