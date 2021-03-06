// file objvalmo.go

package objvalmo  // import "github.com/bstarynk/monimelt/objvalmo"

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"regexp"
	"runtime"
	"serialmo" // import "github.com/bstarynk/monimelt/serialmo"
	"sort"
	"sync"
	"time"
	"unsafe"
)

const (
	TyNilV = iota
	TyIntV
	TyFloatV
	TyStringV
	TyRefobV
	TyColIntV
	TyColStringV
	TyColRefV
	TySetV
	TyTupleV
)

const (
	SpaTransient = iota
	SpaPredefined
	SpaGlobal
	SpaUser
	Spa_Last
)

type ObjectMo struct {
	obid    serialmo.IdentMo
	obmtx   sync.Mutex
	obspace uint8
	obmtime int64
	obattrs map[*ObjectMo]ValueMo
	obcomps []ValueMo
	obpayl  PayloadMo
}

type PayloadMo interface {
	DestroyPayl(pob *ObjectMo)
	DumpScanPayl(pob *ObjectMo, du *DumperMo)
	DumpEmitPayl(pob *ObjectMo, du *DumperMo) (string, interface{})
	GetPayl(pob *ObjectMo, attrpob *ObjectMo) ValueMo
	PutPayl(pob *ObjectMo, attrpob *ObjectMo, val ValueMo) error
	DoPayl(pob *ObjectMo, selpob *ObjectMo, args ...ValueMo) error
} // end PayloadMo

type ValueMo interface {
	TypeV() uint
	Hash() serialmo.HashMo
	DumpScan(*DumperMo)
}

//////////////// string values
type StringVMo interface {
	ValueMo
	isStringV() // private
	Length() int
	ToString() string
}

type StringV struct {
	shash uint32
	str   string
}

func (StringV) isStringV() {}

func (sv StringV) Length() int {
	return len(sv.str)
}

func (sv StringV) ToString() string {
	return sv.str
}

func (sv StringV) TypeV() uint {
	return TyStringV
}

func (sv StringV) DumpScan(du *DumperMo) {}

func StringHash(s string) serialmo.HashMo {
	var h1, h2 uint32
	for ix, ru := range s {
		uc := uint32(ru)
		if ix%2 == 0 {
			h1 = (h1 * 433) ^ ((uc * 1427) + uint32(ix&0xff))
		} else {
			h2 = (h2 * 647) + (uc * 2657) - uint32(ix&0xff)
		}
	}
	h := h1 ^ h2
	if h == 0 {
		h = 3*(h1&0xfffff) + 5*(h2&0xfffff) + uint32(len(s)&0xfffff) + 11
	}
	return serialmo.HashMo(h)
}

func MakeStringV(s string) StringV {
	h := StringHash(s)
	strv := StringV{shash: uint32(h), str: s}
	return strv
}

func (sv StringV) Hash() serialmo.HashMo {
	return serialmo.HashMo(sv.shash)
}

// printable string
func (sv StringV) String() string {
	return fmt.Sprintf("%q", sv.str)
}

//////////////// integer values
type IntVMo interface {
	ValueMo
	isIntV() // private
	Int() int
}

type IntV int

func (IntV) isIntV() {}

func (i IntV) Int() int {
	return int(i)
}

func (IntV) TypeV() uint {
	return TyIntV
}

func (i IntV) Hash() serialmo.HashMo {
	h1 := uint32(i)
	h2 := uint32(i >> 30)
	h := (11 * h1) ^ (26347 * h2)
	if h == 0 {
		h = (h1 & 0xffff) + 17*(h2&0xfffff) + 4
	}
	return serialmo.HashMo(h)
}

func MakeIntV(i int) IntV {
	return IntV(i)
}

func (iv IntV) String() string {
	return fmt.Sprintf("%d", int(iv))
}

func (sv IntV) DumpScan(du *DumperMo) {}

//////////////// float values
type FloatVMo interface {
	ValueMo
	isFloatV() // private
	Float() float64
}

type FloatV float64

func (FloatV) isFloatV() {}
func (f FloatV) Float() float64 {
	return float64(f)
}

func (FloatV) TypeV() uint {
	return TyFloatV
}

func MakeFloatV(f float64) FloatV {
	if math.IsNaN(f) {
		panic("objvalmo.MakeFloatV of NaN")
	}
	return FloatV(f)
}

func (fv FloatV) Hash() serialmo.HashMo {
	f := float64(fv)
	if math.IsNaN(f) {
		panic("objvalmo.Hash of float NaN")
	}
	if math.IsInf(f, 0) {
		if f < 0.0 {
			return 123
		} else {
			return 567
		}
	}
	intp, fracp := math.Modf(f)
	var h uint32
	absintp := math.Abs(intp)
	if absintp < float64(math.MaxInt32) {
		h = uint32(absintp) ^ uint32(fracp*1234567.8)
		if f < 0.0 && h < math.MaxInt32/4 {
			h = 17*h ^ 5023
		}
	} else {
		h = uint32(123.4*math.Log(absintp)) ^ uint32(fracp*456789.0)
		if f < 0.0 && h < math.MaxInt32/4 {
			h = 31*h ^ 15031
		}
	}
	if h == 0 {
		h = ((uint32(math.Log(absintp)) + uint32(fracp*12345678.9)) & 0xfffff) + 17
	}
	return serialmo.HashMo(h)
}
func (fv FloatV) String() string {
	return fmt.Sprintf("%f", float64(fv))
}

func (fv FloatV) DumpScan(du *DumperMo) {}

//////////////// refob values
type RefobVMo interface {
	ValueMo
	isRefobV()
	Obref() *ObjectMo
	IdOb() serialmo.IdentMo
}

type RefobV struct {
	roptr *ObjectMo
}

func (RefobV) isRefobV() {}
func (rob RefobV) TypeV() uint {
	return TyRefobV
}

func (rob RefobV) Obref() *ObjectMo {
	return rob.roptr
}

func (rob RefobV) IdOb() serialmo.IdentMo {
	return rob.roptr.obid
}

func HashObptr(po *ObjectMo) serialmo.HashMo {
	if po == nil {
		return 0
	}
	return po.obid.Hash()
}

func (po *ObjectMo) Hash() serialmo.HashMo {
	return HashObptr(po)
}

func (po *ObjectMo) ToString() string {
	return po.obid.ToString()
}

func LessObptr(pol *ObjectMo, por *ObjectMo) bool {
	if pol == por {
		return false
	}
	if pol == nil {
		return true
	}
	if por == nil {
		return false
	}
	return serialmo.LessId(pol.obid, por.obid)
}

func LessEqualObptr(pol *ObjectMo, por *ObjectMo) bool {
	if pol == por {
		return false
	}
	if pol == nil {
		return true
	}
	if por == nil {
		return false
	}
	return serialmo.LessEqualId(pol.obid, por.obid)
}

func (rob RefobV) Hash() serialmo.HashMo {
	return HashObptr(rob.roptr)
}

func MakeRefobV(pob *ObjectMo) RefobV {
	if pob == nil {
		panic("objectmo.MakeRefobV nil object")
	}
	return RefobV{roptr: pob}
}

func NewRefobV() RefobV {
	pob := NewObj()
	return RefobV{roptr: pob}
}

func RefobSliceToObjptrSlice(rarr []RefobV) []*ObjectMo {
	if rarr == nil {
		return nil
	}
	l := len(rarr)
	osq := make([]*ObjectMo, 0, l)
	for i := 0; i < l; i++ {
		curef := rarr[i]
		if curef.roptr == nil {
			continue
		}
		osq = append(osq, curef.roptr)
	}
	return osq
}

func (rob RefobV) String() string {
	return rob.roptr.obid.ToString()
}

func (rob RefobV) DumpScan(du *DumperMo) {
	du.AddDumpedObject(rob.roptr)
}

//////////////// colored integer values
type ColIntVMo interface {
	ValueMo
	isColIntV()
	ColorRef() *ObjectMo
	Int() int
	Int64() int64
	ColorId() serialmo.IdentMo
}

type ColIntV struct {
	colroptr *ObjectMo
	colint   int64
}

func (ColIntV) isColIntV() {}
func (ci ColIntV) TypeV() uint {
	return TyColIntV
}

func (ci ColIntV) ColorRef() *ObjectMo {
	return ci.colroptr
}

func (ci ColIntV) ColorId() serialmo.IdentMo {
	return ci.colroptr.obid
}

func (ci ColIntV) Int() int     { return int(ci.colint) }
func (ci ColIntV) Int64() int64 { return int64(ci.colint) }
func (ci ColIntV) Hash() serialmo.HashMo {
	var h serialmo.HashMo
	hc := HashObptr(ci.colroptr)
	h = serialmo.HashMo((17 * uint32(hc)) ^ uint32(13*ci.colint))
	if h == 0 {
		h = (hc & 0xffffff) + serialmo.HashMo(ci.colint&0xfffff) + 10
	}
	return h
}

func MakeColInt(colorpob *ObjectMo, num int64) ColIntV {
	if colorpob == nil {
		panic("MakeColInt nil colorpob")
	}
	return ColIntV{colroptr: colorpob, colint: num}
}

func MakeColRefInt(coloref RefobV, num int64) ColIntV {
	return MakeColInt(coloref.Obref(), num)
}

func (ci ColIntV) ToString() string {
	return fmt.Sprintf("%%%s%+d", ci.colroptr.obid.ToString(), ci.colint)
}

func LessColInt(cil ColIntV, cir ColIntV) bool {
	if cil.colroptr == cir.colroptr {
		return cil.colint < cir.colint
	} else {
		return LessObptr(cil.colroptr, cir.colroptr)
	}
}

func LessEqualColInt(cil ColIntV, cir ColIntV) bool {
	if cil.colroptr == cir.colroptr {
		return cil.colint <= cir.colint
	} else {
		return LessObptr(cil.colroptr, cir.colroptr)
	}
}

func (ci ColIntV) DumpScan(du *DumperMo) {
	du.AddDumpedObject(ci.colroptr)
}

//////////////// colored string values
type ColStringVMo interface {
	ValueMo
	isColStringV()
	ColorRef() *ObjectMo
	ColoredString() string
	ColorId() serialmo.IdentMo
}

type ColStringV struct {
	colroptr *ObjectMo
	colstr   string
	colhash  serialmo.HashMo
}

func (ColStringV) isColStringV() {}

func (ci ColStringV) TypeV() uint {
	return TyColStringV
}

func (ci ColStringV) ColorRef() *ObjectMo {
	return ci.colroptr
}

func (ci ColStringV) ColorId() serialmo.IdentMo {
	return ci.colroptr.obid
}

func (ci ColStringV) Hash() serialmo.HashMo {
	return ci.colhash
}

func MakeColString(colorpob *ObjectMo, str string) ColStringV {
	if colorpob == nil {
		panic("MakeColString nil colorpob")
	}
	hs := StringHash(str)
	hc := colorpob.obid.Hash()
	h := (37 * hs) ^ (11 * hc)
	if h == 0 {
		h = (hs & 0xffffff) + 3*(hc&0xffffff) + 10
	}
	return ColStringV{colroptr: colorpob, colstr: str, colhash: h}
}

func MakeColRefStr(coloref RefobV, str string) ColStringV {
	return MakeColString(coloref.Obref(), str)
}

func (ci ColStringV) ToString() string {
	return fmt.Sprintf("%%%s%q", ci.colroptr.obid.ToString(), ci.colstr)
}

func LessColString(cil ColStringV, cir ColStringV) bool {
	if cil.colroptr == cir.colroptr {
		return cil.colstr < cir.colstr
	} else {
		return LessObptr(cil.colroptr, cir.colroptr)
	}
}

func LessEqualColstr(cil ColStringV, cir ColStringV) bool {
	if cil.colroptr == cir.colroptr {
		return cil.colstr <= cir.colstr
	} else {
		return LessObptr(cil.colroptr, cir.colroptr)
	}
}

func (ci ColStringV) DumpScan(du *DumperMo) {
	du.AddDumpedObject(ci.colroptr)
}

//////////////// colored reference values
type ColRefVMo interface {
	ValueMo
	isColRefV()
	ColorRef() *ObjectMo
	ObjRef() *ObjectMo
	ColorId() serialmo.IdentMo
	ObjId() serialmo.IdentMo
}

type ColRefV struct {
	colroptr *ObjectMo
	obroptr  *ObjectMo
}

func (ColRefV) isColRefV() {}

func (ci ColRefV) TypeV() uint {
	return TyColRefV
}

func (ci ColRefV) ColorRef() *ObjectMo {
	return ci.colroptr
}

func (ci ColRefV) ColorId() serialmo.IdentMo {
	return ci.colroptr.obid
}

func (ci ColRefV) ObjRef() *ObjectMo {
	return ci.obroptr
}

func (ci ColRefV) ObjId() serialmo.IdentMo {
	return ci.obroptr.obid
}

func (ci ColRefV) Hash() serialmo.HashMo {
	hc := ci.colroptr.obid.Hash()
	ho := ci.obroptr.obid.Hash()
	h := (47 * hc) ^ (59 * ho)
	if h == 0 {
		h = 3*(hc&0xfffff) + 11*(ho&0xffffff) + 120
	}
	return h
}

func MakeColRef(colorpob *ObjectMo, pob *ObjectMo) ColRefV {
	if colorpob == nil {
		panic("MakeColRef nil colorpob")
	}
	if pob == nil {
		panic("MakeColRef nil pob")
	}
	return ColRefV{colroptr: colorpob, obroptr: pob}
}

func MakeColRefRef(coloref RefobV, oref RefobV) ColRefV {
	return MakeColRef(coloref.Obref(), oref.Obref())
}

func (ci ColRefV) ToString() string {
	return fmt.Sprintf("%%%s/%s", ci.colroptr.obid.ToString(), ci.obroptr.obid.ToString())
}

func LessColRef(cil ColRefV, cir ColRefV) bool {
	if cil.colroptr == cir.colroptr {
		return LessObptr(cil.obroptr, cir.obroptr)
	} else {
		return LessObptr(cil.colroptr, cir.colroptr)
	}
}

func LessEqualColRef(cil ColRefV, cir ColRefV) bool {
	if cil.colroptr == cir.colroptr {
		return LessEqualObptr(cil.obroptr, cir.obroptr)
	} else {
		return LessEqualObptr(cil.colroptr, cir.colroptr)
	}
}

func (ci ColRefV) DumpScan(du *DumperMo) {
	du.AddDumpedObject(ci.colroptr)
	du.AddDumpedObject(ci.obroptr)
}

//////////////// sequence values
type SequenceVMo interface {
	ValueMo
	isSequenceV()         // private
	At(rk int) *ObjectMo  // may panic
	Nth(rk int) *ObjectMo // or nil
	Length() int
}

type SequenceV struct {
	shash  serialmo.HashMo
	scomps []*ObjectMo
}

func (SequenceV) isSequenceV() {}

func (sq SequenceV) At(rk int) *ObjectMo {
	l := len(sq.scomps)
	if rk < 0 {
		rk += l
	}
	if rk < 0 || rk >= l {
		panic("objvalmo.At(SequenceV) out of bounds")
	}
	return sq.scomps[rk]
}

func (SequenceV) TypeV() uint {
	panic("SequenceV.TypeV() impossible")
}
func (sq SequenceV) Length() int {
	return len(sq.scomps)
}
func (sq SequenceV) Hash() serialmo.HashMo { return sq.shash }

func (sq SequenceV) Nth(rk int) *ObjectMo {
	l := len(sq.scomps)
	if rk < 0 {
		rk += l
	}
	if rk < 0 || rk >= l {
		return nil
	}
	return sq.scomps[rk]
}

func (sq SequenceV) DumpScan(du *DumperMo) {
	sln := len(sq.scomps)
	for ix := 0; ix < sln; ix++ {
		curob := sq.scomps[ix]
		du.AddDumpedObject(curob)
	}
}

func (sq SequenceV) seqToString(begc rune, endc rune) string {
	var buf bytes.Buffer
	sln := len(sq.scomps)
	buf.Grow(sln*(2*serialmo.NbDigitsSerialMo+3) + 4)
	buf.WriteRune(begc)
	for ix := 0; ix < sln; ix++ {
		curob := sq.scomps[ix]
		if ix > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(curob.String())
	}
	buf.WriteRune(endc)
	return buf.String()
}

func makeCheckedSequenceSlice(hinit uint32, k1 uint32, k2 uint32, objs []*ObjectMo) SequenceV {
	if objs == nil {
		return SequenceV{}
	}
	l := len(objs)
	var h1, h2 uint32
	h1 = hinit
	h2 = k1*uint32(l) + k2
	sq := make([]*ObjectMo, l)
	for i := 0; i < l; i++ {
		if objs[i] == nil {
			panic("objvalmo.makeCheckedSequence with nil")
		}
		hob := uint32(HashObptr(objs[i]))
		if i%2 == 0 {
			h1 = (k1 * h1) ^ (k2*hob + uint32(i))
		} else {
			h2 = (k2 * h2) + (k1*hob - uint32(5*i))
		}
		sq[i] = objs[i]
	}
	hs := (13 * h1) ^ (4093 * h2)
	if hs == 0 {
		hs = 31*(h1&0xfffff) + 5*(h2&0xfffff) + uint32(17+l&0xff)
	}
	return SequenceV{shash: serialmo.HashMo(hs), scomps: sq}
}

func makeCheckedSequence(hinit uint32, k1 uint32, k2 uint32, objs ...*ObjectMo) SequenceV {
	return makeCheckedSequenceSlice(hinit, k1, k2, objs)
}

func makeSkippedSequenceSlice(hinit uint32, k1 uint32, k2 uint32, objs []*ObjectMo) SequenceV {
	if objs == nil {
		return SequenceV{}
	}
	l := len(objs)
	var h1, h2 uint32
	h1 = hinit
	h2 = k1*uint32(l) + k2
	sq := make([]*ObjectMo, 0, l)
	for i, j := 0, 0; i < l; i++ {
		curobj := objs[i]
		if curobj == nil {
			continue
		}
		hob := uint32(HashObptr(curobj))
		if j%2 == 0 {
			h1 = (k1 * h1) ^ (k2*hob + uint32(j))
		} else {
			h2 = (k2 * h2) + (k1*hob - uint32(5*j))
		}
		sq = append(sq, curobj)
		j++
	}
	hs := (13 * h1) ^ (4093 * h2)
	if hs == 0 {
		hs = 31*(h1&0xfffff) + 5*(h2&0xfffff) + uint32(17+l&0xff)
	}
	return SequenceV{shash: serialmo.HashMo(hs), scomps: sq}
}

func makeSkippedSequence(hinit uint32, k1 uint32, k2 uint32, objs ...*ObjectMo) SequenceV {
	return makeSkippedSequenceSlice(hinit, k1, k2, objs)
}

//////////////// tuple values
type TupleVMo interface {
	SequenceVMo
	isTupleV() // private
}

type TupleV struct {
	SequenceV
}

func (TupleV) TypeV() uint {
	return TyTupleV
}

func (tu TupleV) isTupleV() {}

const hinitTuple = 3529
const k1Tuple = 2521
const k2Tuple = 6529

func MakeTupleV(objs ...*ObjectMo) TupleV {
	return TupleV{makeCheckedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
}

func MakeTupleRefobV(refobjs ...RefobV) TupleV {
	return TupleV{makeCheckedSequenceSlice(hinitTuple, k1Tuple, k2Tuple,
		RefobSliceToObjptrSlice(refobjs))}
}

func MakeTupleSliceV(objs []*ObjectMo) TupleV {
	return TupleV{makeCheckedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
}

func MakeTupleRefobSlice(refobjs []RefobV) TupleV {
	return TupleV{makeCheckedSequenceSlice(hinitTuple, k1Tuple, k2Tuple,
		RefobSliceToObjptrSlice(refobjs))}
}

func MakeSkippedTupleV(objs ...*ObjectMo) TupleV {
	return TupleV{makeSkippedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
}

func MakeSkippedTupleSliceV(objs []*ObjectMo) TupleV {
	return TupleV{makeSkippedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
}

func (tu TupleV) String() string {
	return tu.seqToString('[', ']')
}

// private type for ordering slice of object pointers
type ordSliceObptr []*ObjectMo

func (os ordSliceObptr) Len() int {
	return len(os)
}

func (os ordSliceObptr) Swap(i, j int) {
	os[i], os[j] = os[j], os[i]
}

func (os ordSliceObptr) Less(i, j int) bool {
	return LessObptr(os[i], os[j])
}

func sortedFilteredObptr(arr []*ObjectMo) ordSliceObptr {
	l := len(arr)
	if arr == nil || l == 0 {
		return arr
	}
	coparr := make([]*ObjectMo, 0, l)
	nbnil := 0
	for _, obp := range arr {
		if obp == nil {
			nbnil++
		} else {
			coparr = append(coparr, obp)
		}
	}
	if nbnil == l {
		return ordSliceObptr(nil)
	}
	sort.Sort(ordSliceObptr(coparr))
	coparr = coparr[:l-nbnil]
	hasdup := false
	for ix, ob := range coparr {
		if ix > 0 && ob == coparr[ix-1] {
			hasdup = true
			break
		}
	}
	if !hasdup {
		return ordSliceObptr(coparr)
	}
	resarr := make([]*ObjectMo, 0, l-nbnil)
	resarr = append(resarr, coparr[0])
	for ix, ob := range coparr {
		if ix > 0 && ob != coparr[ix-1] {
			resarr = append(resarr, ob)
		}
	}
	return ordSliceObptr(resarr[:len(resarr)])
}

//////////////// set
type SetVMo interface {
	SequenceVMo
	isSetV() // private
	SetContains(ob *ObjectMo) bool
}

type SetV struct {
	SequenceV
}

func (SetV) TypeV() uint {
	return TySetV
}

func (set SetV) isSetV() {}

const hinitSet = 2549
const k1Set = 3637
const k2Set = 2939

func MakeSetV(objs ...*ObjectMo) SetV {
	ord := sortedFilteredObptr(objs)
	return SetV{makeCheckedSequenceSlice(hinitSet, k1Set, k2Set, ord)}
}

func MakeSetSliceV(objs []*ObjectMo) SetV {
	ord := sortedFilteredObptr(objs)
	return SetV{makeCheckedSequenceSlice(hinitSet, k1Set, k2Set, ord)}
}

func MakeSetRefobV(refobjs ...RefobV) SetV {
	return SetV{makeCheckedSequenceSlice(hinitSet, k1Set, k2Set,
		sortedFilteredObptr(RefobSliceToObjptrSlice(refobjs)))}
}

func MakeSetRefobSlice(refobjs []RefobV) SetV {
	return SetV{makeCheckedSequenceSlice(hinitSet, k1Set, k2Set,
		sortedFilteredObptr(RefobSliceToObjptrSlice(refobjs)))}
}

func (set SetV) String() string {
	return set.seqToString('{', '}')
}

func (set SetV) SetContains(pob *ObjectMo) bool {
	if pob == nil {
		return false
	}
	var lo, hi, md int
	lo = 0
	hi = len(set.scomps)
	for lo+4 < hi {
		md = (lo + hi) / 2
		midpob := set.scomps[md]
		if midpob == pob {
			return true
		}
		if LessObptr(midpob, pob) {
			lo = md
		} else {
			hi = md
		}
	}
	for md = lo; md < hi; md++ {
		midpob := set.scomps[md]
		if midpob == pob {
			return true
		}
	}
	return false
} // end SetContains

////////////////////////////////////////////////////////////////
type bucketTy struct {
	bu_mtx   sync.Mutex
	bu_admap map[serialmo.IdentMo]uintptr
}

var bucketsob [serialmo.MaxBucketMo]bucketTy

func FindObjectById(id serialmo.IdentMo) *ObjectMo {
	if id.EmptyId() {
		return nil
	}
	if !id.ValidId() {
		panic(fmt.Sprintf("objvalmo.FindObjectById invalid id %#x,%#x", id.IdHi, id.IdLo))
	}
	bn := id.BucketNum()
	buck := &bucketsob[bn]
	buck.bu_mtx.Lock()
	defer buck.bu_mtx.Unlock()
	if buck.bu_admap == nil {
		return nil
	}
	ad, ok := buck.bu_admap[id]
	if !ok {
		return nil
	}
	return (*ObjectMo)((unsafe.Pointer)(ad))
}

func FindObjectByTwoNums(hi uint64, lo uint64) *ObjectMo {
	if hi == 0 && lo == 0 {
		return nil
	}
	id := serialmo.IdFromCheckedTwoNums(hi, lo)
	return FindObjectById(id)
}

func finalizeObjectMo(ob *ObjectMo) {
	obid := ob.obid
	ob.obattrs = nil
	ob.obcomps = nil
	ob.obmtime = 0
	p := ob.obpayl
	ob.obpayl = nil
	if p != nil {
		(p).DestroyPayl(ob)
	}
	bn := obid.BucketNum()
	buck := &bucketsob[bn]
	buck.bu_mtx.Lock()
	defer buck.bu_mtx.Unlock()
	delete(buck.bu_admap, obid)
}

func FindOrMakeObjectById(id serialmo.IdentMo) (*ObjectMo, bool) {
	if id.EmptyId() {
		return nil, true
	}
	if !id.ValidId() {
		panic(fmt.Sprintf("objvalmo.FindObjectById invalid id %#x,%#x", id.IdHi, id.IdLo))
	}
	bn := id.BucketNum()
	buck := &bucketsob[bn]
	buck.bu_mtx.Lock()
	defer buck.bu_mtx.Unlock()
	if buck.bu_admap == nil {
		buck.bu_admap = make(map[serialmo.IdentMo]uintptr)
	}
	ad, ok := buck.bu_admap[id]
	if !ok {
		var newobptr *ObjectMo
		newobptr = new(ObjectMo)
		newobptr.obid = id
		buck.bu_admap[id] = uintptr((unsafe.Pointer)(newobptr))
		runtime.SetFinalizer(newobptr, finalizeObjectMo)
		newobptr.UnsyncTouch()
		return newobptr, false
	}
	return (*ObjectMo)((unsafe.Pointer)(ad)), true
}

func MakeObjectById(id serialmo.IdentMo) *ObjectMo {
	ob, _ := FindOrMakeObjectById(id)
	return ob
}

func MakeObjectByTwoNums(hi uint64, lo uint64) *ObjectMo {
	if hi == 0 && lo == 0 {
		return nil
	}
	id := serialmo.IdFromCheckedTwoNums(hi, lo)
	return MakeObjectById(id)
}

func MakePredefinedObj(hi uint64, lo uint64) *ObjectMo {
	pob := MakeObjectByTwoNums(hi, lo)
	pob.UnsyncSetSpaceNum(SpaPredefined)
	log.Printf("MakePredefinedobj pob=%v\n", pob)
	return pob
}

func NewObj() *ObjectMo {
	oid := serialmo.RandomId()
	bn := oid.BucketNum()
	buck := &bucketsob[bn]
	buck.bu_mtx.Lock()
	defer buck.bu_mtx.Unlock()
	if buck.bu_admap == nil {
		buck.bu_admap = make(map[serialmo.IdentMo]uintptr)
	}
	for _, found := buck.bu_admap[oid]; found; oid = serialmo.RandomIdOfBucket(bn) {
	}
	newobptr := new(ObjectMo)
	newobptr.obid = oid
	buck.bu_admap[oid] = uintptr((unsafe.Pointer)(newobptr))
	runtime.SetFinalizer(newobptr, finalizeObjectMo)
	newobptr.UnsyncTouch()
	return newobptr
}

func (pob *ObjectMo) ObId() serialmo.IdentMo {
	if pob == nil {
		return serialmo.TheEmptyId()
	}
	return pob.obid
}

func (pob *ObjectMo) UnsyncTouch() {
	pob.obmtime = time.Now().Unix()
}

func (pob *ObjectMo) UnsyncPutMtime(tim int64) {
	pob.obmtime = tim
}

func (pob *ObjectMo) UnsyncMtime() int64 {
	return pob.obmtime
}

func (pob *ObjectMo) String() string {
	if pob == nil {
		return "__"
	}
	return pob.obid.String()
}

func (pob *ObjectMo) BucketNum() uint {
	if pob == nil {
		panic("objvalmo.BucketNum nil pob")
	}
	return pob.obid.BucketNum()
}

func (pob *ObjectMo) UnsyncSpaceNum() uint8 {
	if pob == nil {
		panic("objvalmo.UnsyncSpaceNum nil pob")
	}
	return pob.obspace
}

func (pob *ObjectMo) SpaceNum() uint8 {
	if pob == nil {
		panic("SpaceNum nil pob")
	}
	pob.obmtx.Lock()
	defer pob.obmtx.Unlock()
	return pob.UnsyncSpaceNum()
}

func (pob *ObjectMo) DumpScanInsideObject(du *DumperMo) {
	log.Printf("DumpScanInsideObject start pob=%v\n", pob)
	defer log.Printf("DumpScanInsideObject end pob=%v\n", pob)
	if pob == nil || du == nil {
		panic("DumpScanInsideObject corruption")
	}
	pob.obmtx.Lock()
	defer pob.obmtx.Unlock()
	log.Printf("DumpScanInsideObject inside pob=%v\n", pob)
	for patob, pval := range pob.obattrs {
		log.Printf("DumpScanInsideObject in pob=%v patob=%v pval=%v\n", pob, patob, pval)
		du.AddDumpedObject(patob)
		if !du.IsDumpedObject(patob) {
			continue
		}
		pval.DumpScan(du)
	}
	for cix, cval := range pob.obcomps {
		log.Printf("DumpScanInsideObject in pob=%v cix=%d cval=%v\n", pob, cix, cval)
		cval.DumpScan(du)
	}
	if pob.obpayl != nil {
		log.Printf("DumpScanInsideObject in pob=%v payload %v\n", pob, pob.obpayl)
		(pob.obpayl).DumpScanPayl(pob, du)
	}
} // end DumpScanInsideObject

var predefined_map map[serialmo.IdentMo]*ObjectMo = make(map[serialmo.IdentMo]*ObjectMo)
var predefined_mtx sync.Mutex

func (pob *ObjectMo) UnsyncSetSpaceNum(sp uint8) *ObjectMo {
	if pob == nil {
		panic("objvalmo.UnsyncSetSpaceNum nil pob")
	}
	if sp >= Spa_Last {
		panic("objvalmo.UnsyncSetSpaceNum out-of-bounds sp")
	}
	oldsp := pob.obspace
	if oldsp == sp {
		return pob
	}
	if oldsp == SpaPredefined {
		predefined_mtx.Lock()
		defer predefined_mtx.Unlock()
		delete(predefined_map, pob.obid)
	}
	if sp == SpaPredefined {
		predefined_mtx.Lock()
		defer predefined_mtx.Unlock()
		predefined_map[pob.obid] = pob
	}
	pob.obspace = sp
	return pob
}

func (pob *ObjectMo) UnsyncPutAttr(pobat *ObjectMo, val ValueMo) *ObjectMo {
	if pob == nil {
		panic("UnsyncPutAttr nil pob")
	}
	if pobat == nil {
		panic(fmt.Errorf("UnsyncPutAttr pob=%v nil pobat", pob))
	}
	if val.TypeV() == TyNilV {
		panic(fmt.Errorf("UnsyncPutAttr pob=%v pobat=%v nil val", pob, pobat))
	}
	if pob.obattrs == nil {
		pob.obattrs = make(map[*ObjectMo]ValueMo)
	}
	pob.obattrs[pobat] = val
	return pob
} // end UnsyncPutAttr

func (pob *ObjectMo) UnsyncAppendVal(val ValueMo) *ObjectMo {
	if pob == nil {
		panic("UnsyncAppendVal nil pob")
	}
	if pob.obcomps == nil {
		pob.obcomps = make([]ValueMo, 0, 7)
	}
	pob.obcomps = append(pob.obcomps, val)
	return pob
} // end UnsyncAppendVal

func (pob *ObjectMo) UnsyncAddValues(vals ...ValueMo) *ObjectMo {
	if pob == nil {
		panic("UnsyncAddValues nil pob")
	}
	if pob.obcomps == nil {
		pob.obcomps = make([]ValueMo, 0, (5+5*len(vals)/4)|7)
	}
	pob.obcomps = append(pob.obcomps, vals...)
	return pob
} // end UnsyncAddValues

func SlicePredefined() []*ObjectMo {
	predefined_mtx.Lock()
	defer predefined_mtx.Unlock()
	nbpr := len(predefined_map)
	sli := make([]*ObjectMo, 0, nbpr)
	for _, pob := range predefined_map {
		sli = append(sli, pob)
	}
	sort.Sort(ordSliceObptr(sli))
	return sli
}

func SetPredefined() SetV {
	return MakeSetSliceV(SlicePredefined())
}

func DumpScanPredefined(du *DumperMo) {
	predefined_mtx.Lock()
	defer predefined_mtx.Unlock()
	for _, pob := range predefined_map {
		du.AddDumpedObject(pob)
	}
}

////////////////////////////////////////////////////////////////
//// global variables support. They should be registered, at init
//// time, using RegisterGlobalVariable. For example:
////    var Glob_foo *ObjectMo
////    RegisterGlobalVariable("foo", &Glob_foo)
//// see generated file globals.go

const glovar_regexp_str = `^[a-zA-Z_][a-zA-Z0-9_]*$`

var glovar_map map[string]**ObjectMo = make(map[string]**ObjectMo)
var glovar_regexp *regexp.Regexp = regexp.MustCompile(glovar_regexp_str)
var glovar_mtx sync.Mutex

func RegisterGlobalVariable(vnam string, advar **ObjectMo) {
	glovar_mtx.Lock()
	defer glovar_mtx.Unlock()
	if glovar_regexp == nil {
		glovar_regexp = regexp.MustCompile(glovar_regexp_str)
	}
	if glovar_map == nil {
		glovar_map = make(map[string]**ObjectMo, 100)
	}
	if !glovar_regexp.MatchString(vnam) {
		panic(fmt.Errorf("RegisterGlobalVariable invalid vnam %q", vnam))
	}
	if advar == nil {
		panic(fmt.Errorf("RegisterGlobalVariable null address for vnam %q", vnam))
	}
	glovar_map[vnam] = advar
	{
		var stabuf [2048]byte
		stalen := runtime.Stack(stabuf[:], true)
		log.Printf("RegisterGlobalVariable glovar_map=%v @%p advar=%p vnam=%v\n...stack:\n%s\n\n\n",
			glovar_map, &glovar_map, advar, vnam, string(stabuf[:stalen]))
	}
}

func UnregisterGlobalVariable(vnam string) {
	if !glovar_regexp.MatchString(vnam) {
		panic(fmt.Errorf("UnregisterGlobalVariable invalid vnam %q", vnam))
	}
	glovar_mtx.Lock()
	defer glovar_mtx.Unlock()
	delete(glovar_map, vnam)
}

func GlobalVariableAddress(vnam string) **ObjectMo {
	if !glovar_regexp.MatchString(vnam) {
		panic(fmt.Errorf("GlobalVariableAddress invalid vnam %q", vnam))
	}
	glovar_mtx.Lock()
	defer glovar_mtx.Unlock()
	vad, _ := glovar_map[vnam]
	return vad
}

func NamesGlobalVariables() []string {
	glovar_mtx.Lock()
	defer glovar_mtx.Unlock()
	ln := len(glovar_map)
	sl := make([]string, 0, ln+1)
	for n, _ := range glovar_map {
		sl = append(sl, n)
	}
	sort.Slice(sl, func(i, j int) bool { return sl[i] < sl[j] })
	return sl
}

func DumpScanGlobalVariables(du *DumperMo) {
	log.Printf("DumpScanGlobalVariables start du=%v\n", du)
	var gcnt int
	glovar_mtx.Lock()
	defer glovar_mtx.Unlock()
	log.Printf("DumpScanGlobalVariables glovar_map=%#v\n", glovar_map)
	for _, av := range glovar_map {
		if *av == nil {
			continue
		}
		log.Printf("DumpScanGlobalVariables av=%v *av=%v\n", av, *av)
		du.AddDumpedObject(*av)
		gcnt++
	}
	if gcnt == 0 {
		log.Printf("DumpScanGlobalVariables did not found any global variables\n")
	}
	log.Printf("DumpScanGlobalVariables end gcnt=%d du=%v\n", gcnt, du)
}

////////////////////////////////////////////////////////////////
//// payload support. They should be registered, at init
//// time, using RegisterPayload. For example:
////    RegisterPayload("symbol", symbol_loader)

const payload_regexp_str = `^[a-zA-Z_][a-zA-Z0-9_]*$`

type PayloadLoaderMo func(pkind string, pob *ObjectMo, ld *LoaderMo, jcont interface{}) PayloadMo

var payload_map map[string]PayloadLoaderMo = make(map[string]PayloadLoaderMo, 100)
var payload_regexp *regexp.Regexp = regexp.MustCompile(payload_regexp_str)
var payload_mtx sync.Mutex

func RegisterPayload(pname string, ploader PayloadLoaderMo) {
	if len(pname) == 0 {
		panic("RegisterPayload empty pname")
	}
	if ploader == nil {
		panic(fmt.Errorf("RegisterPayload nil loader for pname %s", pname))
	}
	payload_mtx.Lock()
	defer payload_mtx.Unlock()
	if payload_regexp == nil {
		payload_regexp = regexp.MustCompile(payload_regexp_str)
	}
	if payload_map == nil {
		payload_map = make(map[string]PayloadLoaderMo)
	}

	if !payload_regexp.MatchString(pname) {
		panic(fmt.Errorf("RegisterPayload invalid pname %q", pname))
	}
	payload_map[pname] = ploader
	{
		var stabuf [2048]byte
		stalen := runtime.Stack(stabuf[:], true)
		log.Printf("RegisterPayload payload_map=%v @%p ploader=%p pname=%v\n...stack:\n%s\n\n\n",
			payload_map, &payload_map, ploader, pname, string(stabuf[:stalen]))
	}
} // end of RegisterPayload

func PayloadLoader(pname string) (PayloadLoaderMo, error) {
	payload_mtx.Lock()
	defer payload_mtx.Unlock()
	pb, ok := payload_map[pname]
	if !ok {
		log.Printf("PayloadLoader unknown pname=%q", pname)
		return nil, fmt.Errorf("unknown PayloadLoader %q", pname)
	}
	return pb, nil
} // end PayloadLoader

func (pob *ObjectMo) UnsyncPayloadClear() *ObjectMo {
	if pob == nil {
		panic("UnsyncPayloadClear nil pob")
	}
	pl := pob.obpayl
	if pl == nil {
		return pob
	}
	pob.obpayl = nil
	(pl).DestroyPayl(pob)
	return pob
} // end UnsyncPayloadClear
