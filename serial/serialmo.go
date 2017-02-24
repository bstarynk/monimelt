// file serialmo.go

package serialmo

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
	return sm == 0 || sm >= MinSerialMo && sm < MaxSerialMo
}

func (sm SerialMo) ToString() string {
	if sm == 0 {
		return "_"
	}
	var buf [NbDigitsSerialMo + 2]byte
	for ix := NbDigitsSerialMo + 1; ix > 0; ix-- {
		d := sm % BaseSerialMo
		buf[ix] = B62DigitsMo[d]
	}
	buf[0] = '_'
	return string(buf)
}
