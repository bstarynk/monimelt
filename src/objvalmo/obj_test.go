// file obj_test.go

package objvalmo

import (
	"fmt"
	"serialmo"
	"testing"
)

func TestFirstObj(t *testing.T) {
	fmt.Printf("TestFirstObj start\n")
	fmt.Printf("nilobj %v\n",
		FindObjectById(serialmo.IdFromCheckedSerials((serialmo.SerialMo)(0),
			(serialmo.SerialMo)(0))))
}

func TestMakeObjs(t *testing.T) {
	fmt.Printf("TestMakeObjs start\n")
	ob1 := NewObj()
	ob2 := NewObj()
	fmt.Printf("*ob1=%v of %T hash %v\n*ob2=%v of %T hash %v\n",
		ob1, ob1, ob1.Hash(), ob2, ob2, ob2.Hash())
	if LessObptr(ob1, ob2) {
		fmt.Printf("ob1=%v is less than ob2=%v\n", ob1, ob2)
	} else {
		fmt.Printf("ob1=%v is greater than ob2=%v\n", ob1, ob2)
	}
	ob3 := NewObj()
	ob4 := NewObj()
	fmt.Printf("*ob3=%v of %T hash %v\n*ob4=%v of %T hash %v\n",
		ob3, ob3, ob3.Hash(), ob4, ob4, ob4.Hash())
	if LessObptr(ob3, ob4) {
		fmt.Printf("ob3=%v is less than ob4=%v\n", ob3, ob4)
	} else {
		fmt.Printf("ob3=%v is greater than ob4=%v\n", ob3, ob4)
	}
}

func TestValues(t *testing.T) {
	fmt.Printf("TestValues start\n")
	fmt.Printf("Nilv %v of hash %v\n",
		GetNilV(), GetNilV().Hash())
	s1 := MakeStringV("abc€")
	s2 := MakeStringV("a\nnewline")
	fmt.Printf("strings s1=%v of hash %v,  s2=%v of hash %v\n",
		s1, s1.Hash(), s2, s2.Hash())
	i1 := MakeIntV(12)
	i2 := MakeIntV(-345)
	fmt.Printf("integers i1=%v of hash %v, i2=%v of hash %v\n",
		i1, i1.Hash(), i2, i2.Hash())
}
