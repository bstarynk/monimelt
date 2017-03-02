// file objvalmo/jsonvalmo.go

package objvalmo

import (
	"encoding/json"
	_ "serialmo"
)

type jsonIdent struct {
	oid string `json:"oid"`
}

type jsonSet struct {
	set []string `json:"set"`
}

type jsonTuple struct {
	tup []string `json:"tup"`
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

// we probably should have some JsonEmitter type....
// having this method....
func ValToJson(vem JsonValEmitterMo, v ValueMo) interface{} {
	var isset bool
	var seqv SequenceV
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
			return fv.Float()
		}
	case TyRefobV:
		{
			obv := v.(RefobV)
			if !vem.EmitObjptr(obv.Obref()) {
				return nil
			}
			obid := obv.IdOb()
			return jsonIdent{oid: obid.ToString()}
		}
	case TySetV:
		isset = true
		seqv = v.(SequenceV)
		fallthrough
	case TyTupleV:
		seqv = v.(SequenceV)
		{
			ls := seqv.Length()
			jseq := make([]string, 0, ls)
			for ix := 0; ix < ls; ix++ {
				curcomp := seqv.At(ix)
				if !vem.EmitObjptr(curcomp) {
					continue
				}
				jseq = append(jseq, curcomp.ToString())
			}
			if isset {
				return jsonSet{set: jseq}
			} else {
				return jsonTuple{tup: jseq}
			}
		}

	}
	panic("objvalmo.ToJson incomplete")
	return nil
}

func OutputJsonValue(vem JsonValEmitterMo, enc *json.Encoder, v ValueMo) {
	enc.Encode(ValToJson(vem, v))
}
