// file obj_test.go

package objvalmo

import (
	"fmt"
	"serialmo"
	"testing"
)

func TestFirstObj(t *testing.T) {
	fmt.Printf("TestFirstObj start\n")
	fmt.Printf("nilobj %v\n",
		FindObjectById(serialmo.IdFromCheckedSerials((serialmo.SerialMo)(0),
			(serialmo.SerialMo)(0))))
}

func TestMakeObjs(t *testing.T) {
	fmt.Printf("TestMakeObjs start\n")
	ob1 := objvalmo.NewObj()
	ob2 := objvalmo.NewObj()
	fmt.Printf("ob1=%v ob2=%v\n", ob1, ob2)
}
