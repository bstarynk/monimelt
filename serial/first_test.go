// file serialmo_testing.go

package serialmo

import (
	"fmt"
	"testing"
)

func TestSerialToString(t *testing.T) {
	s1 := SerialMo(2734358116516558954) // _3fZo81e6aIa
	fmt.Printf("TestSerialToString s1=%d\n", s1)
	fmt.Printf("s1:%s\n", s1.ToString())
}

func TestFromStringSerial(t *testing.T) {
	const s2s = "_4Fgo2LZq1AS" /// 3915796129876347282
	const s2n = 3915796129876347282
	fmt.Printf("TestFromStringSerial s2s=%s s2n=%d=%#x\n", s2s, s2n, s2n);
	s2, e := FromString(s2s)
	fmt.Printf("s2=%d=%x e=%v\n", s2, s2, e);
	fmt.Printf("s2:%s\n", s2.ToString())
}
