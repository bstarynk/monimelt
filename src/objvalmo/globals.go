// file objvalmo/globals.go
// it should be generated
package objvalmo

import "log"

var Glob_the_system *ObjectMo

func init() {
	log.Printf("init of globals.go is registering Glob_the_system@%v\n", &Glob_the_system)
	RegisterGlobalVariable("the_system", &Glob_the_system)
	log.Printf("init of glob_the_system globalnames=%v\n", NamesGlobalVariables())
}
