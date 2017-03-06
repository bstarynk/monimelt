// file objvalmo/jsonvalmo.go

package objvalmo

import (
	"bytes"
	"encoding/json"
	"fmt"
	jason "github.com/antonholmquist/jason"
	"math"
	"serialmo"
	"strconv"
	"strings"
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

type JsonValParserMo interface {
	ParseObjptr(string) (*ObjectMo, error)
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

func EncodeJsonValue(vem JsonValEmitterMo, enc *json.Encoder, v ValueMo) {
	enc.Encode(ValToJson(vem, v))
}

func EmitJsonValueInBuffer(vem JsonValEmitterMo, buf *bytes.Buffer, v ValueMo) {
	enc := json.NewEncoder(buf)
	EncodeJsonValue(vem, enc, v)
}

type JsonSimpleValParser struct {
}

func (JsonSimpleValParser) ParseObjptr(sid string) (*ObjectMo, error) {
	oid, err := serialmo.IdFromString(sid)
	if err != nil {
		return nil, err
	}
	pob, ok := FindOrMakeObjectById(oid)
	if !ok {
		return nil, fmt.Errorf("JsonSimpleValParser.ParseObjptr bad sid=%q", sid)
	}
	return pob, nil
}

func TrivialValParser() JsonSimpleValParser {
	return JsonSimpleValParser{}
}

func JasonParseValue(vpm JsonValParserMo, jval jason.Value) (ValueMo, error) {
	var err error
	//fmt.Printf("JasonParseValue start jval %v (%T)\n", jval, jval)
	err = jval.Null()
	//fmt.Printf("JasonParseValue jval %v (%T) err=%v\n", jval, jval, err)
	if err == nil {
		return nil, nil
	} else if num, err := jval.Number(); err == nil {
		ns := num.String()
		if strings.ContainsRune(ns, '.') || strings.ContainsRune(ns, 'e') || strings.ContainsRune(ns, 'e') {
			fv, _ := num.Float64()
			return MakeFloatV(fv), nil
		} else {
			iv, _ := num.Int64()
			return MakeIntV(int(iv)), nil
		}
	} else if str, err := jval.String(); err == nil {
		return MakeStringV(str), nil
	} else {
		job, err := jval.Object()
		if err != nil {
			return nil, err
		}
		if obs, err := job.GetString("oid"); err == nil {
			pob, err := vpm.ParseObjptr(obs)
			if pob != nil && err == nil {
				return MakeRefobV(pob), nil
			} else if err != nil {
				return nil, err
			} else {
				return nil, nil
			}
		} else if flos, err := job.GetString("float"); err == nil {
			if flos == "+Inf" {
				return MakeFloatV(math.Inf(+1)), nil
			} else if flos == "-Inf" {
				return MakeFloatV(math.Inf(-1)), nil
			}
		} else if oelems, err := job.GetStringArray("set"); err == nil {
			l := len(oelems)
			obseq := make([]*ObjectMo, 0, l)
			for ix := 0; ix < l; ix++ {
				pob, err := vpm.ParseObjptr(oelems[ix])
				if pob != nil && err == nil {
					obseq = append(obseq, pob)
				} else if err != nil {
					return nil, err
				}
			}
			return MakeSetSliceV(obseq), nil
		} else if ocomps, err := job.GetStringArray("tup"); err == nil {
			l := len(ocomps)
			obseq := make([]*ObjectMo, 0, l)
			for ix := 0; ix < l; ix++ {
				pob, err := vpm.ParseObjptr(ocomps[ix])
				if pob != nil && err == nil {
					obseq = append(obseq, pob)
				} else if err != nil {
					return nil, err
				}
			}
			return MakeTupleSliceV(obseq), nil
		} else if jval, err := job.GetValue("value"); err == nil {
			return JasonParseValue(vpm, *jval)
		}
	}
	return nil, fmt.Errorf("JasonParseValue invalid jval: %v", jval)
}
