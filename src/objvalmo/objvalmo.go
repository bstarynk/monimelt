// file objvalmo.go

package objvalmo

import (
	"fmt"
	"math"
	"runtime"
	"serialmo"
	"sort"
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

func (h HashMo) String() string {
	return fmt.Sprintf("h#%d", uint32(h))
}

type ValueMo interface {
	TypeV() uint
	Hash() HashMo
}

//////////////// the nil value

type NilVMo interface {
	ValueMo
	isNilV()
}

type NilV struct {
}

var nilValue = NilV{}

func (NilV) String() string { return "__" }
func (NilV) Hash() HashMo   { return HashMo(0) }
func (NilV) isNilV()        {}

func GetNilV() NilV { return nilValue }

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

func (sv StringV) ToString() string {
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

func (sv StringV) Hash() HashMo {
	return HashMo(sv.shash)
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

func (iv IntV) String() string {
	return fmt.Sprintf("%d", int(iv))
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
func (fv FloatV) String() string {
	return fmt.Sprintf("%f", float64(fv))
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

func (po *ObjectMo) Hash() HashMo {
	return HashObptr(po)
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

func NewRefobV() RefobV {
	pob := NewObj()
	return RefobV{roptr: pob}
}
func (rob RefobV) String() string {
	return rob.roptr.obid.ToString()
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

func (sq SequenceV) seqToString(begc rune, endc rune) string {
	panic("objvalmo.seqToString unimplemented")
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
	return SequenceV{shash: HashMo(hs), scomps: sq}
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

func MakeTupleSliceV(objs []*ObjectMo) TupleV {
	return TupleV{makeCheckedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
}

func MakeSkippedTupleV(objs ...*ObjectMo) TupleV {
	return TupleV{makeSkippedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
}

func MakeSkippedTupleSliceV(objs []*ObjectMo) TupleV {
	return TupleV{makeSkippedSequenceSlice(hinitTuple, k1Tuple, k2Tuple, objs)}
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
	return newobptr
}

func (pob *ObjectMo) String() string {
	if pob == nil {
		return "__"
	}
	return pob.obid.String()
}
