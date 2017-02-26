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
		panic(fmt.Sprintf("FromCheckedUint64 %#x fail %v",
			u, e))
	}
	return sr
}
