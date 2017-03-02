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
	s1 := MakeStringV("abc€")
	s2 := MakeStringV("a\nnewline")
	fmt.Printf("strings s1=%v of hash %v,  s2=%v of hash %v\n",
		s1, s1.Hash(), s2, s2.Hash())
	i1 := MakeIntV(12)
	i2 := MakeIntV(-345)
	fmt.Printf("integers i1=%v of hash %v, i2=%v of hash %v\n",
		i1, i1.Hash(), i2, i2.Hash())
	f1 := MakeFloatV(12.3)
	f2 := MakeFloatV(-1.0)
	f3 := MakeFloatV(11.0e20)
	f3bis := MakeFloatV(11.0e20)
	fmt.Printf("floats f1=%v of hash %v, f2=%v of hash %v, f3=%v of hash %v, f3bis=%v of hash %v\n", f1, f1.Hash(), f2, f2.Hash(), f3, f3.Hash(), f3bis, f3bis.Hash())
	ro1 := NewRefobV()
	ro2 := NewRefobV()
	ro3 := NewRefobV()
	fmt.Printf("refobjs ro1=%v of hash %v, ro2=%v of hash %v, ro3=%v of hash %v\n",
		ro1, ro1.Hash(), ro2, ro2.Hash(), ro3, ro3.Hash())

	tu1 := MakeTupleRefobV(ro1, ro2, ro3, ro2, ro1)
	tu2 := MakeSkippedTupleV(ro1.Obref(), nil, ro2.Obref(), nil, ro3.Obref())
	fmt.Printf("tuples tu1=%v of hash %v, tu2=%v of hash %v\n",
		tu1, tu1.Hash(), tu2, tu2.Hash())
	ro4 := NewRefobV()
	ro5 := NewRefobV()
	fmt.Printf("refobjs ro4=%v of hash %v, ro5=%v of hash %v\n",
		ro4, ro4.Hash(), ro5, ro5.Hash())
	set1 := MakeSetRefobV(ro1, ro2, ro3, ro4, ro5)
	set1bis := MakeSetRefobV(ro5, ro4, ro3, ro2, ro1, ro5)
	fmt.Printf("set set1=%v of hash %v, set1bis=%v of hash %v\n",
		set1, set1.Hash(), set1bis, set1bis.Hash())
	set2 := MakeSetRefobV(ro1, ro2)
	set3 := MakeSetRefobV(ro3, ro3, ro4, ro2)
	fmt.Printf("set set2=%v of hash %v, set3=%v of hash %v\n",
		set2, set2.Hash(), set3, set3.Hash())
	skipobmap := make(map[*ObjectMo]*ObjectMo)
	jsem := MakeJsonSimpleValEmitter(func(pob *ObjectMo) bool {
		_, ok := skipobmap[pob]
		if ok {
			return true
		}
		bn := pob.BucketNum()
		if bn%2 == 0 {
			skipobmap[pob] = pob
			fmt.Printf("simplejsonemitter adding pob %v of bucket#%d\n",
				pob, bn)
			return true
		} else {
			return false
		}
	})
	fmt.Printf("jsem %v of type %T\n", jsem, jsem)
}
