// file values-mo.go
// manage values
package valuesmo

type ValMo interface {
	isValueMo()
}

type IntMo int
func (i IntMo) isValueMo() {}

type StringMo string
func (s StringMo) isValueMo() {}
