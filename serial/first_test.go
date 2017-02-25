// file serialmo_testing.go

package serialmo

import (
	"fmt"
	"testing"
)

func TestSerialToString(t *testing.T) {
	s1 := SerialMo(2734358116516558954) // _3fZo81e6aIa
	fmt.Printf("s1=%d\n", s1)
	fmt.Printf("s1:%s\n", s1.ToString())
}
