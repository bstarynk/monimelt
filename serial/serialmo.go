// file serialmo.go

package serialmo

import (
	"errors"
	"strings"
)

type SerialMo uint64

const B62DigitsMo = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const MinSerialMo uint64 = 62 * 62
const MaxSerialMo uint64 = 10 * 62 * (62 * 62 * 62) * (62 * 62 * 62) * (62 * 62 * 62)
const DeltaSerialMo uint64 = MaxSerialMo - MinSerialMo
const NbDigitsSerialMo = 11
const BaseSerialMo = 62

type IdentMo struct {
	IdHi, IdLo SerialMo
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
	var buf [NbDigitsSerialMo + 2]byte
	for ix := NbDigitsSerialMo + 1; ix > 0; ix-- {
		d := sm % BaseSerialMo
		sm = sm / BaseSerialMo
		buf[ix] = B62DigitsMo[d]
	}
	buf[0] = '_'
	return string(buf[:])
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
	for ix := NbDigitsSerialMo + 1; ix > 0; ix-- {
		c := s[ix]
		r := strings.IndexByte(B62DigitsMo, c)
		if r < 0 {
			return SerialMo(0), errors.New("serialmo.FromString invalid char")
		}
		sr = sr*SerialMo(BaseSerialMo) + SerialMo(r)
	}
	return sr, nil
}
