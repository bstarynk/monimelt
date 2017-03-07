// file objvalmo/globals.go
// it should be generated
package objvalmo

var Glob_the_system *ObjectMo

func init() {
	RegisterGlobalVariable("the_system", &Glob_the_system)
}
