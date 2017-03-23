// file serialmo.go

package serialmo

import (
	cryptrand "crypto/rand"
	encbinary "encoding/binary"
	"errors"
	"fmt"
	mathrand "math/rand"
	"strings"
)

type SerialMo uint64

const B62DigitsMo = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const MinSerialMo uint64 = 62 * 62
const MaxSerialMo uint64 = 10 * 62 * (62 * 62 * 62) * (62 * 62 * 62) * (62 * 62 * 62)
const DeltaSerialMo uint64 = MaxSerialMo - MinSerialMo
const NbDigitsSerialMo = 11
const BaseSerialMo = 62
const MaxBucketMo = 10 * 62

type HashMo uint32

func (h HashMo) String() string {
	return fmt.Sprintf("h#%d", uint32(h))
}

type IdentMo struct {
	IdHi, IdLo SerialMo
}

var randchan chan uint64

const seedperiod = 8192

func randroutine() {
	var randcount uint64
	var curseed int64
	for {
		if randcount%seedperiod == 0 {
			var buf [8]byte
			cryptrand.Read(buf[:])
			curseed = int64(encbinary.LittleEndian.Uint64(buf[:]))
			mathrand.Seed(curseed)
		}
		randcount++
		randchan <- mathrand.Uint64()
	}
}

func init() {
	randchan = make(chan uint64)
	go randroutine()
}

func (sm SerialMo) ValidSerial() bool {
	return uint64(sm) == 0 ||
		uint64(sm) >= MinSerialMo && uint64(sm) < MaxSerialMo
}

func (sm SerialMo) NonEmpty() bool {
	return uint64(sm) != 0
}

func (sm SerialMo) Empty() bool {
	return uint64(sm) == 0
}

func (sm SerialMo) ToString() string {
	if sm == 0 {
		return "_"
	}
	var buf [NbDigitsSerialMo + 1]byte
	for ix := NbDigitsSerialMo; ix > 0; ix-- {
		d := sm % BaseSerialMo
		sm = sm / BaseSerialMo
		buf[ix] = B62DigitsMo[d]
	}
	buf[0] = '_'
	return string(buf[:])
}

func (sm SerialMo) BucketNum() uint {
	return uint(uint64(sm) / uint64(DeltaSerialMo/MaxBucketMo))
}

func (sm SerialMo) BucketOffset() uint64 {
	return uint64(sm) % (DeltaSerialMo / MaxBucketMo)
}

func RandomSerial() SerialMo {
	var r uint64
	for r < MinSerialMo || r >= MaxSerialMo {
		r = <-randchan
	}
	return SerialMo(r)
}

func (sm SerialMo) ToUint64() uint64 {
	return uint64(sm)
}

func RandomOfBucket(bn uint) SerialMo {
	if bn >= MaxBucketMo {
		panic(fmt.Sprintf("serialmo.RandomOfBucket bad bn=%d", bn))
	}
	r := <-randchan % (DeltaSerialMo / MaxBucketMo)
	s := (uint64(bn) * (DeltaSerialMo / MaxBucketMo)) + r + MinSerialMo
	return SerialMo(s)
}

func FromString(s string) (SerialMo, error) {
	if s == "" {
		return SerialMo(0), errors.New("serialmo.FromString empty string")
	}
	if s[0] != '_' {
		return SerialMo(0), errors.New("serialmo.FromString string does not start with underscore")
	}
	if len(s) != NbDigitsSerialMo+1 {
		return SerialMo(0), errors.New("serialmo.FromString string of wrong length")
	}
	sr := SerialMo(0)
	for ix := 1; ix <= NbDigitsSerialMo; ix++ {
		c := s[ix]
		r := strings.IndexByte(B62DigitsMo, c)
		if r < 0 {
			return SerialMo(0), errors.New("serialmo.FromString invalid char")
		}
		sr = sr*SerialMo(BaseSerialMo) + SerialMo(r)
	}
	return sr, nil
}

func FromCheckedString(s string) SerialMo {
	sr, e := FromString(s)
	if e != nil {
		panic(fmt.Sprintf("serialmo.FromCheckedString %s fail %v",
			s, e))
	}
	return sr
}

func FromUint64(u uint64) (SerialMo, error) {
	if u == 0 {
		return SerialMo(0), nil
	}
	if u < MinSerialMo || u >= MaxSerialMo {
		return SerialMo(0), errors.New("serialmo.FromUint64 out of bound")
	}
	return SerialMo(u), nil
}

func FromCheckedUint64(u uint64) SerialMo {
	sr, e := FromUint64(u)
	if e != nil {
		panic(fmt.Sprintf("serialmo.FromCheckedUint64 %#x fail %v",
			u, e))
	}
	return sr
}

func (id IdentMo) EmptyId() bool {
	return uint64(id.IdHi) == 0 && uint64(id.IdLo) == 0
}

func TheEmptyId() IdentMo {
	return IdentMo{IdHi: 0, IdLo: 0}
}

func (id IdentMo) ValidId() bool {
	if id.EmptyId() {
		return true
	}
	return id.IdHi.ValidSerial() && id.IdLo.ValidSerial()
}

func (id IdentMo) ToString() string {
	if id.EmptyId() {
		return "__"
	}
	return id.IdHi.ToString() + id.IdLo.ToString()
}

func (id IdentMo) String() string {
	return id.ToString()
}

func (id IdentMo) Hash() HashMo {
	if id.EmptyId() {
		return 0
	}
	h := uint32((id.IdHi * 1033) ^ (id.IdLo * 2027))
	if h == 0 {
		h = (uint32(id.IdHi) & 0xfffff) + 17*(uint32(id.IdLo)&0xfffff) + 30
	}
	return HashMo(h)
}

func (id IdentMo) BucketNum() uint {
	return id.IdHi.BucketNum()
}

func (id IdentMo) ToTwoNums() (uint64, uint64) {
	return uint64(id.IdHi), uint64(id.IdLo)
}

func LessId(idl IdentMo, idr IdentMo) bool {
	if idl.IdHi < idr.IdHi {
		return true
	}
	if idl.IdHi > idr.IdHi {
		return false
	}
	if idl.IdLo == idr.IdLo {
		return false
	}
	if idl.IdLo < idr.IdLo {
		return true
	}
	return false
}

func LessEqualId(idl IdentMo, idr IdentMo) bool {
	if idl.IdHi < idr.IdHi {
		return true
	}
	if idl.IdHi > idr.IdHi {
		return false
	}
	if idl.IdLo == idr.IdLo {
		return true
	}
	if idl.IdLo < idr.IdLo {
		return true
	}
	return false
}

func RandomId() IdentMo {
	return IdentMo{IdHi: RandomSerial(), IdLo: RandomSerial()}
}

func RandomIdOfBucket(bn uint) IdentMo {
	return IdentMo{IdHi: RandomOfBucket(bn), IdLo: RandomSerial()}
}

func IdFromString(s string) (IdentMo, error) {
	if s == "" {
		return IdentMo{}, errors.New("serialmo.IdFromString empty string")
	}
	if s == "__" {
		return IdentMo{}, nil
	}
	if len(s) != 2*NbDigitsSerialMo+2 {
		return IdentMo{}, errors.New("serialmo.IdFromString string of wrong length")
	}
	if s[0] != '_' {
		return IdentMo{}, errors.New("serialmo.IdFromString string does not start with underscore")
	}
	hi, eh := FromString(s[0 : NbDigitsSerialMo+1])
	if eh != nil {
		return IdentMo{}, errors.New("serialmo.IdFromString bad hi part")
	}
	lo, el := FromString(s[NbDigitsSerialMo+1 : 2*NbDigitsSerialMo+2])
	if el != nil {
		return IdentMo{}, errors.New("serialmo.IdFromString bad lo part")
	}
	return IdentMo{IdHi: hi, IdLo: lo}, nil
}

func IdFromCheckedString(s string) IdentMo {
	id, e := IdFromString(s)
	if e != nil {
		panic(fmt.Sprintf("serialmo.IdFromCheckedString failure %v", e))
	}
	return id
}

func IdFromSerials(shi SerialMo, slo SerialMo) (IdentMo, error) {
	if uint64(shi) == 0 && uint64(slo) == 0 {
		return IdentMo{}, nil
	}
	if uint64(shi) == 0 {
		return IdentMo{}, errors.New("serialmo.IdFromSerials zero shi")
	}
	if uint64(slo) == 0 {
		return IdentMo{}, errors.New("serialmo.IdFromSerials zero slo")
	}
	if !shi.ValidSerial() {
		return IdentMo{}, errors.New("serialmo.IdFromSerials invalid shi")
	}
	if !slo.ValidSerial() {
		return IdentMo{}, errors.New("serialmo.IdFromSerials invalid slo")
	}
	return IdentMo{IdHi: shi, IdLo: slo}, nil
}

func IdFromCheckedSerials(shi SerialMo, slo SerialMo) IdentMo {
	id, e := IdFromSerials(shi, slo)
	if e != nil {
		panic(fmt.Sprintf("serialmo.IdFromCheckedSerials failure %v", e))
	}
	return id
}

func IdFromCheckedTwoNums(nhi uint64, nlo uint64) IdentMo {
	id, e := IdFromSerials(SerialMo(nhi), SerialMo(nlo))
	if e != nil {
		panic(fmt.Sprintf("serialmo.IdFromCheckedTwoNums failure %v", e))
	}
	return id
}
