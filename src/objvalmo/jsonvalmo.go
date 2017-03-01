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

// we probably should have some JsonEmitter type....
// having this method....
func ValToJson(v ValueMo) interface{} {
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
			obid := obv.IdOb()
			return jsonIdent{oid: obid.ToString()}
		}
	}
	panic("objvalmo.ToJson unimplemented")
	return nil
}

func OutputJsonValue(enc *json.Encoder, v ValueMo) {
	enc.Encode(ValToJson(v))
}
