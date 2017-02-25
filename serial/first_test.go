// file serialmo_testing.go

package serialmo

import (
	"fmt"
	"testing"
)

func TestSerialToString(t *testing.T) {
	s1 := SerialMo(2734358116516558954) // _3fZo81e6aIa
	fmt.Printf("s1=%d:%s\n", s1, s1.ToString())
}
