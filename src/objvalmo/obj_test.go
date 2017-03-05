// file obj_test.go

package objvalmo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jason"
	"math"
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

func json_emit(jem JsonSimpleValEmitter, msg string, v ValueMo) {
	fmt.Printf("json_emit %s v=%v of type %T\n", msg, v, v)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	EncodeJsonValue(jem, enc, v)
	fmt.Printf("json_emit %s buf: %s\n", msg, buf.String())
}

func json_parse(msg string, js string) {
	fmt.Printf("json_parse %s: %s\n", msg, js)
	jv, err := jason.NewValueFromBytes(([]byte)(js))
	if err != nil {
		fmt.Printf("json_parse jason failure %s: %v\n", msg, err)
		return
	}
	fmt.Printf("json_parse %s: jv %v // %T\n", msg, jv, jv)
	tp := TrivialValParser()
	v, err := JasonParseValue(tp, *jv)
	if err != nil {
		fmt.Printf("json_parse failure %s: %v\n\n", msg, err)
	} else {
		fmt.Printf("json_parse success %s: %v (%T)\n\n", msg, v, v)
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
	f1 := MakeFloatV(-12.3)
	f2 := MakeFloatV(-1.0)
	f3 := MakeFloatV(11.0e20)
	f3bis := MakeFloatV(11.0e20)
	fmt.Printf("floats f1=%v of hash %v, f2=%v of hash %v, f3=%v of hash %v, f3bis=%v of hash %v\n", f1, f1.Hash(), f2, f2.Hash(), f3, f3.Hash(), f3bis, f3bis.Hash())
	f4 := MakeFloatV(math.Pi)
	f5 := MakeFloatV(math.E * 1.0e150)
	fmax := MakeFloatV(math.MaxFloat64)
	finf := MakeFloatV(math.Inf(+1))
	fmt.Printf("floats f4=%v of hash %v, f5=%v of hash %v\n",
		f4, f4.Hash(), f5, f5.Hash())
	fmt.Printf("floats fmax=%v of hash %v, finf=%v of hash %v\n",
		fmax, fmax.Hash(), finf, finf.Hash())
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
	json_emit(jsem, "i1", i1)
	json_emit(jsem, "i2", i2)
	json_emit(jsem, "f1", f1)
	json_emit(jsem, "f2", f2)
	json_emit(jsem, "f3", f3)
	json_emit(jsem, "f4", f4)
	json_emit(jsem, "f5", f5)
	json_emit(jsem, "fmax", fmax)
	json_emit(jsem, "finf", finf)
	json_emit(jsem, "ro1", ro1)
	json_emit(jsem, "ro2", ro2)
	json_emit(jsem, "ro3", ro3)
	json_emit(jsem, "ro4", ro4)
	json_emit(jsem, "ro5", ro5)
	json_emit(jsem, "tu1", tu1)
	json_emit(jsem, "tu2", tu2)
	json_emit(jsem, "set1", set1)
	json_emit(jsem, "set2", set2)
	json_emit(jsem, "set3", set3)
	///
	jv, err := jason.NewValueFromBytes(([]byte)("null"))
	fmt.Printf("objtest 'null' jv=%v (%T) err=%v\n", jv, jv, err)
	json_parse("test-nil", "null") /// for some reason, test-nil is failing...
	json_parse("test-valuenil", `{"value":null}`)
	json_parse("test-1", " 1")
	json_parse("test-m23", "-23")
}
