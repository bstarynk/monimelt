// file objvalmo.go

package objvalmo

import (
	"fmt"
	"math"
	"runtime"
	"serialmo"
	"sync"
	"unsafe"
)

const (
	TyNilV = iota
	TyIntV
	TyFloatV
	TyStringV
	TyRefobV
	TySetV
	TyTupleV
)

type ObjectMo struct {
	obid    serialmo.IdentMo
	obmtx   sync.Mutex
	obspace uint8
	obattrs map[*ObjectMo]ValueMo
	obcomps []ValueMo
	obpayl  *PayloadMo
}

type PayloadMo interface {
	DestroyPayl(*ObjectMo)
}

type HashMo uint32

type ValueMo interface {
	TypeV() uint
	Hash() HashMo
}

//////////////// string values
type StringVMo interface {
	ValueMo
	isStringV() // private
	Length() int
	String() string
}

type StringV struct {
	shash uint32
	str   string
}

func (StringV) isStringV() {}

func (sv StringV) Length() int {
	return len(sv.str)
}

func (sv StringV) String() string {
	return sv.str
}

func (sv StringV) TypeV() uint {
	return TyStringV
}

func StringHash(s string) HashMo {
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
	return HashMo(h)
}

func MakeStringV(s string) StringV {
	h := StringHash(s)
	strv := StringV{shash: uint32(h), str: s}
	return strv
}

func Hash(sv StringV) HashMo {
	return HashMo(sv.shash)
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

func (i IntV) Hash() HashMo {
	h1 := uint32(i)
	h2 := uint32(i >> 30)
	h := (11 * h1) ^ (26347 * h2)
	if h == 0 {
		h = (h1 & 0xffff) + 17*(h2&0xfffff) + 4
	}
	return HashMo(h)
}

func MakeIntV(i int) IntV {
	return IntV(i)
}

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

func (fv FloatV) Hash() HashMo {
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
	return HashMo(h)
}

//////////////// refob values
type RefobVMo interface {
	ValueMo
	isRefobV()
	Obref() *ObjectMo
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

func HashObptr(po *ObjectMo) HashMo {
	if po == nil {
		return 0
	}
	nhi, nlo := po.obid.ToTwoNums()
	h := uint32((nhi * 1033) ^ (nlo * 2027))
	if h == 0 {
		h = (uint32(nhi) & 0xfffff) + 17*(uint32(nlo)&0xfffff) + 30
	}
	return HashMo(h)
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

func (rob RefobV) Hash() HashMo {
	return HashObptr(rob.roptr)
}

func MakeRefobV(pob *ObjectMo) RefobV {
	if pob == nil {
		panic("objectmo.MakeRefobV nil object")
	}
	return RefobV{roptr: pob}
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
	shash  HashMo
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

func (sq SequenceV) Length() int {
	return len(sq.scomps)
}
func (sq SequenceV) Hash() HashMo { return sq.shash }

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
	return SequenceV{shash: HashMo(hs), scomps: sq}
}

func makeCheckedSequence(hinit uint32, k1 uint32, k2 uint32, objs ...*ObjectMo) SequenceV {
	return makeCheckedSequenceSlice(hinit, k1, k2, objs)
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
	var tup TupleV
	tup = TupleV{makeCheckedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
	return tup
}

func MakeTupleSliceV(objs []*ObjectMo) TupleV {
	var tup TupleV
	tup = TupleV{makeCheckedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
	return tup
}

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
	p := ob.obpayl
	ob.obpayl = nil
	if p != nil {
		(*p).DestroyPayl(ob)
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
	ad, ok := buck.bu_admap[id]
	if !ok {
		var newobptr *ObjectMo
		newobptr = new(ObjectMo)
		newobptr.obid = id
		buck.bu_admap[id] = uintptr((unsafe.Pointer)(newobptr))
		runtime.SetFinalizer(*newobptr, finalizeObjectMo)
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