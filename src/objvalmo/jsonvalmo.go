// file objvalmo/jsonvalmo.go

package objvalmo

import (
	"encoding/json"
	"fmt"
	"math"
	_ "serialmo"
	"strconv"
)

type jsonIdent struct {
	Joid string `json:"oid"`
}

type jsonSet struct {
	Jset []string `json:"set"`
}

type jsonTuple struct {
	Jtup []string `json:"tup"`
}
type JsonValEmitterMo interface {
	EmitObjptr(*ObjectMo) bool
}

type EmitterFunction_t func(*ObjectMo) bool

type JsonSimpleValEmitter struct {
	emfun EmitterFunction_t
}

func MakeJsonSimpleValEmitter(ef EmitterFunction_t) JsonSimpleValEmitter {
	return JsonSimpleValEmitter{emfun: ef}
}

func (jse JsonSimpleValEmitter) EmitObjptr(pob *ObjectMo) bool {
	return jse.emfun(pob)
}

/// see https://groups.google.com/forum/#!topic/golang-nuts/nIshrMRrAt0
type myJsonFloat float64

func (mf myJsonFloat) MarshalJSON() ([]byte, error) {
	f := float64(mf)
	if math.IsNaN(f) {
		return []byte("null"), nil
	}
	if math.IsInf(f, +1) {
		return []byte(`{"float":"+Inf"}`), nil
	}
	if math.IsInf(f, -1) {
		return []byte(`{"float":"-Inf"}`), nil
	}
	s := fmt.Sprintf("%.4f", f)
	x := math.NaN()
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q first\n", f, x, s)
	if x == f && len(s) < 20 {
		return []byte(s), nil
	}
	s = fmt.Sprintf("%.9f", f)
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q second/f\n", f, x, s)
	if x == f && len(s) < 25 {
		return []byte(s), nil
	}
	s = fmt.Sprintf("%.9e", f)
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q second/e\n", f, x, s)
	if x == f {
		return []byte(s), nil
	}
	s = fmt.Sprintf("%.15e", f)
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q third\n", f, x, s)
	if x == f {
		return []byte(s), nil
	}
	s = fmt.Sprintf("%.28E", f)
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q last\n", f, x, s)
	return []byte(s), nil
}

func (fv FloatV) MarshalJSON() ([]byte, error) {
	fmt.Printf("FloatV MarshalJSON fv=%v\n", fv)
	return myJsonFloat(fv.Float()).MarshalJSON()
}

func sequenceToJsonTuple(vem JsonValEmitterMo, seqv SequenceV) []string {
	ls := seqv.Length()
	jseq := make([]string, 0, ls)
	for ix := 0; ix < ls; ix++ {
		curcomp := seqv.At(ix)
		if !vem.EmitObjptr(curcomp) {
			continue
		}
		jseq = append(jseq, curcomp.ToString())
	}
	return jseq
}

// we probably should have some JsonEmitter type....
// having this method....
func ValToJson(vem JsonValEmitterMo, v ValueMo) interface{} {
	switch v.TypeV() {
	case TyIntV:
		{
			iv := v.(IntV)
			return iv.Int()
		}
	case TyStringV:
		{
			sv := v.(StringV)
			return sv.ToString()
		}
	case TyFloatV:
		{
			fv := v.(FloatV)
			return myJsonFloat(fv.Float())
		}
	case TyRefobV:
		{
			obv := v.(RefobV)
			if !vem.EmitObjptr(obv.Obref()) {
				return nil
			}
			obid := obv.IdOb()
			return jsonIdent{Joid: obid.ToString()}
		}
	case TySetV:
		{
			setv := v.(SetV)
			return jsonSet{Jset: sequenceToJsonTuple(vem, setv.SequenceV)}
		}
	case TyTupleV:
		{
			tupv := v.(TupleV)
			return jsonTuple{Jtup: sequenceToJsonTuple(vem, tupv.SequenceV)}
		}
	}
	panic("objvalmo.ToJson incomplete")
	return nil
}

func OutputJsonValue(vem JsonValEmitterMo, enc *json.Encoder, v ValueMo) {
	enc.Encode(ValToJson(vem, v))
}
