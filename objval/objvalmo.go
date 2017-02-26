// file objvalmo.go

package objvalmo

import (
	"fmt"
	serialmo "github.com/bstarynk/monimelt/serial"
	"runtime"
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
	obattrs map[*ObjectMo]*ValueMo
	obcomps []*ValueMo
	obpayl  *PayloadMo
}

type PayloadMo interface {
	DestroyPayl(*ObjectMo)
}

type ValueMo interface {
}

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

func MakeObjectById(id serialmo.IdentMo) *ObjectMo {
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
		var newobptr *ObjectMo
		newobptr = new(ObjectMo)
		newobptr.obid = id
		buck.bu_admap[id] = uintptr((unsafe.Pointer)(newobptr))
		runtime.SetFinalizer(*newobptr, finalizeObjectMo)
		return newobptr
	}
	return (*ObjectMo)((unsafe.Pointer)(ad))
}
