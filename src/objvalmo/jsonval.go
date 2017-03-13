// file objvalmo/jsonval.go

package objvalmo

import (
	"bytes"
	"encoding/json"
	"fmt"
	jason "github.com/antonholmquist/jason"
	"log"
	"math"
	"serialmo"
	"strconv"
	"strings"
)

type jsonIdent struct {
	Joid string `json:"oid"`
}

type jsonInt struct {
	Jint string `json:"int"`
}

type jsonFloat struct {
	Jfloat string `json:"float"`
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
	x := math.NaN()
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
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q first\n", f, x, s)
	if x == f && len(s) < 20 {
		return []byte(`{"float":"` + s + `"}`), nil
	}
	s = fmt.Sprintf("%.9f", f)
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q second/f\n", f, x, s)
	if x == f && len(s) < 25 {
		return []byte(`{"float":"` + s + `"}`), nil
	}
	s = fmt.Sprintf("%.9e", f)
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q second/e\n", f, x, s)
	if x == f {
		return []byte(`{"float":"` + s + `"}`), nil
	}
	s = fmt.Sprintf("%.15e", f)
	x, _ = strconv.ParseFloat(s, 64)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q third\n", f, x, s)
	if x == f {
		return []byte(`{"float":"` + s + `"}`), nil
	}
	s = fmt.Sprintf("%.28E", f)
	//fmt.Printf("myJsonFloat f=%f x=%f s=%q last\n", f, x, s)
	return []byte(`{"float":"` + s + `"}`), nil
}

func (fv FloatV) MarshalJSON() ([]byte, error) {
	fmt.Printf("FloatV MarshalJSON fv=%v\n", fv)
	return myJsonFloat(fv.Float()).MarshalJSON()
}

func (iv IntV) MarshalJSON() ([]byte, error) {
	log.Printf("IntV MarshalJSON iv=%v\n", iv)
	i := iv.Int()
	if i > -1000000000 && i < 1000000000 {
		return ([]byte)(fmt.Sprintf("%d", i)), nil
	} else {
		return ([]byte)(fmt.Sprintf(`{"int":"%d"}`, i)), nil
	}
} // end IntV MarshalJSON

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
	var res interface{}
	log.Printf("ValToJson v=%#v (%T)\n", v, v)
	defer log.Printf("ValToJson v=%v (%T) res=%v (%T)\n", v, v, res, res)
	switch v.TypeV() {
	case TyIntV:
		{
			iv := v.(IntV)
			i := iv.Int()
			if i > -1000000000 && i < 1000000000 {
				res = i
			} else {
				res = jsonInt{Jint: fmt.Sprintf("%d", i)}
			}
			return res
		}
	case TyStringV:
		{
			sv := v.(StringV)
			res = sv.ToString()
			return res
		}
	case TyFloatV:
		{
			fv := v.(FloatV)
			res = myJsonFloat(fv)
			return res
		}
	case TyRefobV:
		{
			obv := v.(RefobV)
			if !vem.EmitObjptr(obv.Obref()) {
				return nil
			}
			obid := obv.IdOb()
			res = jsonIdent{Joid: obid.ToString()}
			return res
		}
	case TySetV:
		{
			setv := v.(SetV)
			res = jsonSet{Jset: sequenceToJsonTuple(vem, setv.SequenceV)}
			return res
		}
	case TyTupleV:
		{
			tupv := v.(TupleV)
			res = jsonTuple{Jtup: sequenceToJsonTuple(vem, tupv.SequenceV)}
			return res
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

func JasonParseVal(vpm JsonValParserMo, jv interface{}) (ValueMo, error) {
	var resval ValueMo
	var err error
	log.Printf("JasonParseVal start jv %#v (%T)\n", jv, jv)
	defer log.Printf("JasonParseVal end jv %#v (%T) resval %#v (%T) err %v\n\n", jv, jv, resval, resval, err)
	if jv == nil {
		resval = nil
		return resval, nil
	} else if jstr, ok := jv.(string); ok {
		log.Printf("JasonParseVal jstr=%q\n", jstr)
		resval = MakeStringV(jstr)
		log.Printf("JasonParseVal string resval=%#v (%T)\n", resval, resval)
		return resval, nil
	} else if jint, ok := jv.(int); ok {
		log.Printf("JasonParseVal jint=%d\n", jint)
		resval = MakeIntV(jint)
		return resval, nil
	} else if jintl, ok := jv.(int64); ok {
		log.Printf("JasonParseVal jintl=%d\n", jintl)
		resval = MakeIntV(int(jintl))
		return resval, nil
	} else if jflo, ok := jv.(float64); ok {
		log.Printf("JasonParseVal jflo=%g (%T)\n", jflo, jflo)
		resval = MakeFloatV(jflo)
		return resval, nil
	} else if jmap, ok := jv.(map[string]interface{}); ok {
		log.Printf("JasonParseVal jmap %#v (%T)\n", jmap, jmap)
		if joid, ok := jmap["oid"]; ok {
			pob, err := vpm.ParseObjptr(joid.(string))
			if pob != nil && err == nil {
				resval = MakeRefobV(pob)
				return resval, nil
			}
		} else if jflos, ok := jmap["float"]; ok {
			var flos string
			log.Printf("JasonParseVal jflos=%#v (%T)\n", jflos, jflos)
			if flos, ok = jflos.(string); !ok {
				err = fmt.Errorf("JasonParseVal bad float %#v (%T)", jflos, jflos)
				return nil, err
			}
			if flos == "+Inf" {
				resval = MakeFloatV(math.Inf(+1))
				return resval, nil
			} else if flos == "-Inf" {
				resval = MakeFloatV(math.Inf(-1))
				return resval, nil
			} else {
				fnum, err := strconv.ParseFloat(flos, 64)
				if err != nil {
					return nil, err
				}
				resval = MakeFloatV(fnum)
				return resval, nil
			}
		} else if jints, ok := jmap["int"]; ok {
			var intstr string
			log.Printf("JasonParseVal jints=%#v (%T)\n", jints, jints)
			if intstr, ok = jints.(string); !ok {
				err = fmt.Errorf("JasonParseVal bad jints %#v (%T)", jints, jints)
				return nil, err
			}
			intnum, err := strconv.ParseInt(intstr, 0, 64)
			if err != nil {
				return nil, err
			}
			resval = MakeIntV(int(intnum))
			return resval, nil
		} else if jelemset, ok := jmap["set"]; ok {
			if jelems, ok := jelemset.([]string); ok {
				l := len(jelems)
				obseq := make([]*ObjectMo, 0, l)
				for ix := 0; ix < l; ix++ {
					pob, err := vpm.ParseObjptr(jelems[ix])
					if pob != nil && err == nil {
						obseq = append(obseq, pob)
					} else if err != nil {
						return nil, err
					}
				}
				resval = MakeSetSliceV(obseq)
				return resval, nil
			}
		} else if jcomptup, ok := jmap["tup"]; ok {
			if jcomps, ok := jcomptup.([]string); ok {
				l := len(jcomps)
				obseq := make([]*ObjectMo, 0, l)
				for ix := 0; ix < l; ix++ {
					pob, err := vpm.ParseObjptr(jcomps[ix])
					if pob != nil && err == nil {
						obseq = append(obseq, pob)
					} else if err != nil {
						return nil, err
					}
				}
				resval = MakeTupleSliceV(obseq)
				return resval, nil
			}
		}
		err = fmt.Errorf("JasonParseVal unexpected jmap %#v (%T)", jmap, jmap)
		return nil, err
	}
	jval, ok := jv.(jason.Value)
	if !ok {
		err = fmt.Errorf("JasonParseVal invalid jv %v (%T) not jason.Value", jv, jv)
		return resval, err
	}
	err = jval.Null()
	//fmt.Printf("JasonParseVal jval %v (%T) err=%v\n", jval, jval, err)
	if err == nil {
		resval = nil
		return resval, nil
	} else if num, err := jval.Number(); err == nil {
		ns := num.String()
		if strings.ContainsRune(ns, '.') || strings.ContainsRune(ns, 'e') || strings.ContainsRune(ns, 'e') {
			fv, _ := num.Float64()
			resval = MakeFloatV(fv)
			return resval, nil
		} else {
			iv, _ := num.Int64()
			resval = MakeIntV(int(iv))
			return resval, nil
		}
	} else if str, err := jval.String(); err == nil {
		resval = MakeStringV(str)
		return resval, nil
	} else {
		job, err := jval.Object()
		if err != nil {
			return nil, err
		}
		if obs, err := job.GetString("oid"); err == nil {
			pob, err := vpm.ParseObjptr(obs)
			if pob != nil && err == nil {
				resval = MakeRefobV(pob)
				return resval, nil
			} else if err != nil {
				return nil, err
			} else {
				resval = nil
				return resval, nil
			}
		} else if flos, err := job.GetString("float"); err == nil {
			if flos == "+Inf" {
				resval = MakeFloatV(math.Inf(+1))
				return resval, nil
			} else if flos == "-Inf" {
				resval = MakeFloatV(math.Inf(-1))
				return resval, nil
			}
			fnum, err := strconv.ParseFloat(flos, 64)
			if err != nil {
				return nil, err
			}
			resval = MakeFloatV(fnum)
			return resval, nil
		} else if intstr, err := job.GetString("int"); err == nil {
			intnum, err := strconv.ParseInt(intstr, 0, 64)
			if err != nil {
				return nil, err
			}
			resval = MakeIntV(int(intnum))
			return resval, nil
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
			resval = MakeSetSliceV(obseq)
			return resval, nil
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
			resval = MakeTupleSliceV(obseq)
			return resval, nil
		} else if jval, err := job.GetValue("value"); err == nil {
			return JasonParseVal(vpm, *jval)
		}
	}
	return nil, fmt.Errorf("JasonParseVal invalid jval: %v (%T)", jval, jval)
}
