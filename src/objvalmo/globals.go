// file objvalmo/globals.go
// it should be generated
package objvalmo

import "log"

var Glob_the_system *ObjectMo

func init() {
	log.Printf("init of globals.go in registering Glob_the_system\n")
	RegisterGlobalVariable("the_system", &Glob_the_system)
}
