// file objvalmo.go

package objvalmo

import (
	serialmo "github.com/bstarynk/monimelt/serial"
	"unsafe"
	"sync"
	"fmt"
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
	obid serialmo.IdentMo
	obattrs map[*ObjectMo]*ValueMo
	obcomps []*ValueMo
	obpayl *PayloadMo
}

type PayloadMo interface {
}

type ValueMo interface {
}

type bucketTy struct {
	bu_mtx sync.Mutex
	bu_admap map[serialmo.IdentMo] uintptr
}

var bucketsob [serialmo.MaxBucketMo]bucketTy

func FindObjectById (id serialmo.IdentMo) *ObjectMo {
	if (id.EmptyId()) {
		return nil
	}
	if (!id.ValidId()) {
		panic (fmt.Sprintf("objvalmo.FindObjectById invalid id %#x,%#x", id.IdHi, id.IdLo))
	}
	bn := id.BucketNum()
	buck := &bucketsob[bn]
	buck.bu_mtx.Lock()
	defer buck.bu_mtx.Unlock()
	ad, ok :=  buck.bu_admap[id]
	if (!ok) {
		return nil
	}
	return (*ObjectMo)((unsafe.Pointer)(ad))
}


