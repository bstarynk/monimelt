// file monimelt/monimelt.go
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"serialmo"
)

func main() {
	hasSerialPtr := flag.Bool("serial", false, "generate serials and obids")
	nbSerialPtr := flag.Int("nb-serial", 2, "number of serials")

	flag.Parse()
	fmt.Printf("Monimelt starting pid %d, Go version %s\n", os.Getpid(), runtime.Version())
	if *hasSerialPtr {
		n := *nbSerialPtr
		fmt.Printf("Monimelt %d serials\n", n)
		for i := 0; i < n; i++ {
			sr := serialmo.RandomSerial()
			fmt.Printf("serial#%d: %d=%#x %s buck#%d offset:%d\n",
				i, sr.ToUint64(), sr.ToUint64(), sr.ToString(), sr.BucketNum(), sr.BucketOffset())
		}
		fmt.Printf("Monimelt %d objids\n", n)
		for i := 0; i < n; i++ {
			oid := serialmo.RandomId()
			nhi, nlo := oid.ToTwoNums()
			fmt.Printf("id#%d: %#x,%#x %s buck#%d\n",
				i, nhi, nlo, oid.ToString(), oid.BucketNum())
		}
	}
}
