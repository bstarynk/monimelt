// file obj_test.go

package objvalmo

import (
	"fmt"
	serialmo "github.com/bstarynk/monimelt/serial"
	"testing"
)

func TestFirstObj(t *testing.T) {
	fmt.Printf("TestFirstObj start\n")
	fmt.Printf("nilobj %v\n",
		FindObjectById(serialmo.IdFromCheckedSerials((serialmo.SerialMo)(0),
			(serialmo.SerialMo)(0))))
}
