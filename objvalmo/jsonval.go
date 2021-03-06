// file objvalmo/jsonval.go

package objvalmo  // import "github.com/bstarynk/monimelt/objvalmo"

import (
	"bytes"
	"encoding/json"
	"fmt"
	jason "github.com/antonholmquist/jason"
	"log"
	"math"
	"strconv"
	"strings"
	"serialmo" // import "github.com/bstarynk/monimelt/serialmo"
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

type jsonColInt struct {
	Jcolori  int64  `json:"colori"`
	Jcolorob string `json:"colorob"`
}

type jsonColString struct {
	Jcolorstr string `json:"colorstr"`
	Jcolorob  string `json:"colorob"`
}

type jsonColRef struct {
	Jcoloref string `json:"coloref"`
	Jcolorob string `json:"colorob"`
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
	vty := v.TypeV()
	log.Printf("ValToJson v=%#v vty:%d (%T)\n", v, vty, v)
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
	case TyColIntV:
		{
			civ := v.(ColIntV)
			if !vem.EmitObjptr(civ.ColorRef()) {
				return nil
			}
			cobid := civ.ColorId()
			res = jsonColInt{Jcolori: civ.colint, Jcolorob: cobid.ToString()}
			return res
		}
	case TyColStringV:
		{
			csv := v.(ColStringV)
			if !vem.EmitObjptr(csv.ColorRef()) {
				return nil
			}
			cobid := csv.ColorId()
			res = jsonColString{Jcolorstr: csv.colstr, Jcolorob: cobid.ToString()}
			return res
		}
	case TySetV:
		{
			setv := v.(SetV)
			res = jsonSet{Jset: sequenceToJsonTuple(vem, setv.SequenceV)}
			log.Printf("ValToJson setv=%v res=%v\n", setv, res)
			return res
		}
	case TyTupleV:
		{
			tupv := v.(TupleV)
			res = jsonTuple{Jtup: sequenceToJsonTuple(vem, tupv.SequenceV)}
			log.Printf("ValToJson tupv=%v res=%v\n", tupv, res)
			return res
		}
	}
	panic(fmt.Errorf("objvalmo.ToJson incomplete v=%v", v))
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
	defer log.Printf("JasonParseVal end jv %v (%T) resval %#v (%T) err %v\n\n", jv, jv, resval, resval, err)
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
		var intnum int
		intnum = int(jflo)
		if float64(intnum) == jflo {
			resval = MakeIntV(intnum)
		} else {
			resval = MakeFloatV(jflo)
		}
		return resval, nil
	} else if jmap, ok := jv.(map[string]interface{}); ok {
		log.Printf("JasonParseVal jmap %#v (%T)\n", jmap, jmap)
		//// object reference: {"oid":....}
		if joid, ok := jmap["oid"]; ok {
			pob, err := vpm.ParseObjptr(joid.(string))
			if pob != nil && err == nil {
				resval = MakeRefobV(pob)
				return resval, nil
			}
		} else
		//// floating point value: {"float": ...}
		if jflos, ok := jmap["float"]; ok {
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
		} else
		//// integer value: {"int": ...}
		if jints, ok := jmap["int"]; ok {
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
		} else
		//// set value: {"set": [ ... ] }
		if jelemset, ok := jmap["set"]; ok {
			log.Printf("JasonParseVal set jv=%v jelemset=%v (%T)",
				jv, jelemset, jelemset)
			if jelems, ok := jelemset.([]interface{}); ok {
				l := len(jelems)
				obseq := make([]*ObjectMo, 0, l)
				for ix := 0; ix < l; ix++ {
					jcurelemstr, ok := jelems[ix].(string)
					if !ok {
						err = fmt.Errorf("JasonParseVal bad jelemset %#v (%T) ix=%d", jelemset, jelemset, ix)
						log.Printf("JasonParseVal set!err=%v jv=%v\n",
							err, jv)
						return nil, err

					}
					pob, err := vpm.ParseObjptr(jcurelemstr)
					if pob != nil && err == nil {
						obseq = append(obseq, pob)
					} else if err != nil {
						return nil, err
					}
				}
				resval = MakeSetSliceV(obseq)
				log.Printf("JasonParseVal set resval=%v jv=%v\n",
					resval, jv)
				return resval, nil
			} else {
				err = fmt.Errorf("JasonParseVal bad jelemset %#v (%T)", jelemset, jelemset)
				log.Printf("JasonParseVal set!err=%v jv=%v\n",
					err, jv)
				return nil, err
			}
		} else
		//// tuple value: { "tup" : [ ... ] }
		if jcomptup, ok := jmap["tup"]; ok {
			log.Printf("JasonParseVal tup jv=%v jcomptup=%v (%T)",
				jv, jcomptup, jcomptup)
			if jcomps, ok := jcomptup.([]interface{}); ok {
				l := len(jcomps)
				obseq := make([]*ObjectMo, 0, l)
				for ix := 0; ix < l; ix++ {
					jcurcompstr, ok := jcomps[ix].(string)
					if !ok {
						err = fmt.Errorf("JasonParseVal bad jcomptup %#v (%T) ix=%d", jcomptup, jcomptup, ix)
						log.Printf("JasonParseVal tup!err=%v jv=%v\n",
							err, jv)
						return nil, err
					}
					pob, err := vpm.ParseObjptr(jcurcompstr)
					if pob != nil && err == nil {
						obseq = append(obseq, pob)
					} else if err != nil {
						return nil, err
					}
				}
				resval = MakeTupleSliceV(obseq)
				log.Printf("JasonParseVal tup resval=%v jv=%v\n",
					resval, jv)
				return resval, nil
			} else {
				err = fmt.Errorf("JasonParseVal bad jcomptup %#v (%T)", jcomptup, jcomptup)
				return nil, err
			}
		} else
		/// colored integer value { "colori": <int> ; "colorob" : <obref> }
		if jcolori, ok := jmap["colori"]; ok {
			var ic int64
			var colpob *ObjectMo
			if fic, ok := jcolori.(float64); ok {
				ic = int64(fic)
			} else if lic, ok := jcolori.(int64); ok {
				ic = lic
			} else if nic, ok := jcolori.(int); ok {
				ic = int64(nic)
			} else if sic, ok := jcolori.(string); ok {
				if inum, err := strconv.ParseInt(sic, 0, 64); err == nil {
					ic = inum
				} else {
					return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorint %v", jmap, err)
				}
			} else {
				return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorint %v (strange \"colori\")", jmap, err)
			}
			if jcid, ok := jmap["colorob"]; ok {
				if jcidstr, ok := jcid.(string); ok {
					if colpob, err = vpm.ParseObjptr(jcidstr); err != nil {
						return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorint wrong \"colorob\" %v", jmap, err)
					}
				} else {
					return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorint strange \"colorob\"", jmap)
				}

			} else {
				return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorint missing \"colorob\"", jmap)
			}
			resval = MakeColInt(colpob, ic)
			return resval, nil
		} else
		/// colored string value { "colorstr": <string> ; "colorob" : <obref> }
		if jcolorstr, ok := jmap["colorstr"]; ok {
			var colstr string
			var colpob *ObjectMo
			if str, ok := jcolorstr.(string); ok {
				colstr = str
			} else {
				return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorstr %v (strange \"colorstr\")", jmap, jcolorstr)
			}
			if jcid, ok := jmap["colorob"]; ok {
				if jcidstr, ok := jcid.(string); ok {
					if colpob, err = vpm.ParseObjptr(jcidstr); err != nil {
						return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorstr wrong \"colorob\" %v", jmap, err)
					}
				} else {
					return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorstr strange \"colorob\"", jmap)
				}

			} else {
				return nil, fmt.Errorf("JasonParseVal jmap %#v bad colorstr missing \"colorob\"", jmap)
			}
			resval = MakeColString(colpob, colstr)
			return resval, nil
		} else
		//// colored reference value { "coloref" : <reference> , "colorob" : <color> }
		if jcoloref, hascolorefok := jmap["coloref"]; hascolorefok {
			var colorefids string
			var colorobids string
			var colerr error
			var jcolorob interface{}
			var refpob *ObjectMo
			var colpob *ObjectMo
			var strcolorefok bool
			var strcolorobok bool
			var hascolorobok bool
			if colorefids, strcolorefok = jcoloref.(string); !strcolorefok {
				colerr = fmt.Errorf("bad coloref %v", jcoloref)
				goto badcoloref
			}
			if jcolorob, hascolorobok = jmap["colorob"]; !hascolorobok {
				colerr = fmt.Errorf("no colorob")
				goto badcoloref
			}
			if colorobids, strcolorobok = jcolorob.(string); !strcolorobok {
				colerr = fmt.Errorf("bad colorob %v", jcolorob)
				goto badcoloref
			}
			if refpob, colerr = vpm.ParseObjptr(colorefids); colerr != nil {
				colerr = fmt.Errorf("invalid coloref %v", colerr)
				goto badcoloref
			}
			if colpob, colerr = vpm.ParseObjptr(colorobids); colerr != nil {
				colerr = fmt.Errorf("invalid colorob %v", colerr)
				goto badcoloref
			}
			resval = MakeColRef(colpob, refpob)
			return resval, nil
		badcoloref:
			return nil, fmt.Errorf("bad coloref %v : %v", jcoloref, colerr)
		}
		//// otherwise, error
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
