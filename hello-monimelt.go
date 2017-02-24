// file monimelt/hello-monimelt.go
package main

import "fmt"
import "runtime"

func main() {
	fmt.Printf("hello, world from monimelt, Go version %s\n", runtime.Version())
}
