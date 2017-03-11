// file monimelt/monimelt.go
package main

import (
	"flag"
	"fmt"
	"log"
	"objvalmo"
	"os"
	"runtime"
	"serialmo"
	"time"
)

func main() {
	hasSerialPtr := flag.Bool("serial", false, "generate serials and obids")
	nbSerialPtr := flag.Int("nb-serial", 3, "number of serials")
	loadPtr := flag.String("load", "", "initial load directory")
	tinyDump1Ptr := flag.String("tiny-dump1", "", "directory to dump with DoTinyDump1")
	finalDumpPtr := flag.String("final-dump", "", "final dump directory")
	flag.Parse()
	log.Printf("Monimelt starting pid %d, Go version %s\n", os.Getpid(), runtime.Version())
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
			fmt.Printf("id#%d: %#x,%#x %s buck#%d %v\n",
				i, nhi, nlo, oid.ToString(), oid.BucketNum(), oid.Hash())
		}
	}
	//
	if len(*loadPtr) > 0 {
		log.Printf("monimelt should initial load from %s\n", *loadPtr)
		objvalmo.LoadFromDirectory(*loadPtr)
		log.Printf("monimelt did initial load from %s\n", *loadPtr)
	}
	//
	time.Sleep(30 * time.Millisecond)
	if len(*tinyDump1Ptr) > 0 {
		log.Printf("monimelt should dotinydump1 in %s\n", *tinyDump1Ptr)
		objvalmo.DoTinyDump1(*tinyDump1Ptr)
		log.Printf("monimelt did dotinydump1 in %s\n", *tinyDump1Ptr)

	}
	//
	if len(*finalDumpPtr) > 0 {
		log.Printf("monimelt should final dump in %s\n", *finalDumpPtr)
		objvalmo.DumpIntoDirectory(*finalDumpPtr)
		log.Printf("monimelt did final dump in %s\n", *finalDumpPtr)
	}
	log.Printf("Monimelt ending pid %d\n", os.Getpid())
}
