// file objvalmo.go

package objvalmo

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"regexp"
	"runtime"
	"serialmo"
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
	obpayl  *PayloadMo
}

type PayloadMo interface {
	DestroyPayl(*ObjectMo)
	DumpScanPayl(*ObjectMo, *DumperMo)
	DumpEmitPayl(*ObjectMo, *DumperMo) (string, interface{})
}

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
	if pob == nil || du == nil {
		panic("DumpScanInsideObject corruption")
	}
	pob.obmtx.Lock()
	defer pob.obmtx.Unlock()
	for patob, pval := range pob.obattrs {
		du.AddDumpedObject(patob)
		if !du.IsDumpedObject(patob) {
			continue
		}
		pval.DumpScan(du)
	}
	for _, cval := range pob.obcomps {
		cval.DumpScan(du)
	}
	if pob.obpayl != nil {
		(*pob.obpayl).DumpScanPayl(pob, du)
	}
}

var predefined_map map[serialmo.IdentMo]*ObjectMo
var predefined_mtx sync.Mutex

func init() {
	predefined_map = make(map[serialmo.IdentMo]*ObjectMo)
}

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

var glovar_map map[string]**ObjectMo
var glovar_regexp *regexp.Regexp
var glovar_mtx sync.Mutex

const glovar_regexp_str = `^[a-zA-Z_][a-zA-Z0-9_]*$`

func init() {
	glovar_map = make(map[string]**ObjectMo)
	glovar_regexp = regexp.MustCompile(glovar_regexp_str)
}

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
		log.Printf("RegisterGlobalVariable glovar_map=%v advar=%v vnam=%v\n...stack:\n%s\n\n\n",
			glovar_map, advar, vnam, string(stabuf[:stalen]))
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
