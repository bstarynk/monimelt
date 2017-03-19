/// file makename.go
package main

import (
	"log"
	"objvalmo"
	"payloadmo"
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
