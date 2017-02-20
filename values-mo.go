// file values-mo.go
// manage values
package values_mo

type ValMo interface {
	isValueMo()
}

type IntMo int
func (i IntMo) isValueMo() {}

type StringMo string
func (s StringMo) isValueMo() {}
