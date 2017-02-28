// file objvalmo.go

package objvalmo

import (
	"fmt"
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
	obattrs map[*ObjectMo]*ValueMo
	obcomps []*ValueMo
	obpayl  *PayloadMo
}

type PayloadMo interface {
	DestroyPayl(*ObjectMo)
}

type ValueMo interface {
	TypeV() uint
}

//////////////// string values
type StringVMo interface {
	ValueMo
	isStringV() // private
	Length() int
	String() string
}

type StringV string

func (StringV) isStringV() {}

func (sv StringV) Length() int {
	return len(sv)
}

func (sv StringV) String() string {
	return string(sv)
}

func (sv StringV) TypeV() uint {
	return TyStringV
}

func MakeStringV(s string) StringV {
	return StringV(s)
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

func MakeFloatV(f float64) FloatV {
	return FloatV(f)
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

//////////////// sequence values
type SequenceVMo interface {
	ValueMo
	isSequenceV()         // private
	At(rk int) *ObjectMo  // may panic
	Nth(rk int) *ObjectMo // or nil
	Length() int
}

type SequenceV struct {
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

func makeCheckedSequence(objs ...*ObjectMo) SequenceV {
	if objs == nil {
		return SequenceV{}
	}
	l := len(objs)
	sq := make([]*ObjectMo, l)
	for i := 0; i < l; i++ {
		if objs[i] == nil {
			panic("objvalmo.makeCheckedSequence with nil")
		}
		sq[i] = objs[i]
	}
	return SequenceV{scomps: sq}
}

//////////////// tuple values
type TupleVMo interface {
	SequenceVMo
	isTupleV() // private
}

type TupleV struct {
	SequenceV
}

func (tu TupleV) isTupleV() {}

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
