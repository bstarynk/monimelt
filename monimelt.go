// file monimelt/monimelt.go
package main  // import "github.com/bstarynk/monimelt/"

import (
	"flag"
	"fmt"
	"log"
	"os"
	osexec "os/exec"
	"path"
	"plugin"
	"runtime"
	"time"
	/// our packages:
	"objvalmo" // import "github.com/bstarynk/monimelt/objvalmo"
	"serialmo" // import "github.com/bstarynk/monimelt/serialmo"
	_ "payloadmo" // import "github.com/bstarynk/monimelt/payloadmo"
)

func main() {
	hasSerialPtr := flag.Bool("serial", false, "generate serials and obids")
	nbSerialPtr := flag.Int("nb-serial", 3, "number of serials")
	loadPtr := flag.String("load", "", "initial load directory")
	tinyDump1Ptr := flag.String("tiny-dump1", "", "directory to dump with DoTinyDump1")
	pluginRunPtr := flag.String("run-plugin", "", "Go source file to compile and load as plugin")
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
		time.Sleep(10 * time.Millisecond)
		log.Printf("monimelt should dotinydump1 in %s\n", *tinyDump1Ptr)
		objvalmo.DoTinyDump1(*tinyDump1Ptr)
		log.Printf("monimelt did dotinydump1 in %s\n", *tinyDump1Ptr)

	}
	//
	if len(*pluginRunPtr) > 3 {
		var pluginsrc string
		var err error
		var cmd *osexec.Cmd
		var plug *plugin.Plugin
		var symb plugin.Symbol
		pluginsrc = *pluginRunPtr
		time.Sleep(10 * time.Millisecond)
		log.Printf("monimelt should run as plugin %s\n", pluginsrc)
		lenplugin := len(pluginsrc)
		sharedpath := path.Base(pluginsrc[0:lenplugin-3]) + ".so"
		log.Printf("monimelt pluginsrc %q sharedpath %q", pluginsrc, sharedpath)
		if _, err := os.Stat(sharedpath); err == nil {
			if err = os.Remove(sharedpath); err != nil {
				log.Printf("failed to remove shared %q - %v\n", sharedpath, err)
				goto pluginend
			} else {
				log.Printf("did remove shared %q\n", sharedpath)
			}
		}
		if pluginsrc[lenplugin-3:] != ".go" {
			log.Printf("plugin %q not ending with .go\n", pluginsrc)
		}
		if _, err := os.Stat(pluginsrc); err != nil {
			log.Printf("missing plugin source %s - %v", pluginsrc, err)
			goto pluginend
		}
		cmd = osexec.Command("./build-plugin-monimelt.sh", pluginsrc)
		log.Printf("plugin cmd=%v\n", cmd)
		if err = cmd.Run(); err != nil {
			log.Printf("plugin %s build failure; %v\n", pluginsrc, err)
			goto pluginend
		}
		if _, err := os.Stat(sharedpath); err == nil {
			log.Printf("plugin will open sharedpath %s\n", sharedpath)
		} else {
			log.Printf("plugin without sharedpath %s - %v\n", sharedpath, err)
			goto pluginend
		}
		if plug, err = plugin.Open(sharedpath); err != nil {
			log.Printf("plugin %s open %s failure: %v\n", sharedpath, err)
			goto pluginend
		}
		log.Printf("plugin %s opened %v\n", sharedpath, plug)
		if symb, err = plug.Lookup("DoMonimelt"); err == nil {
			log.Printf("plugin %s has 'DoMonimelt' %v\n", sharedpath, symb)
			if fsy, ok := symb.(func()); ok {
				log.Printf("before running fsy %#v from plugin %s\n", fsy, sharedpath)
				fsy()
				log.Printf("after running fsy %#v from plugin %s\n", fsy, sharedpath)
			} else {
				log.Printf("plugin %s has strange 'DoMonimelt' %v\n", sharedpath, fsy)
				goto pluginend
			}
		}
		log.Printf("done plugin %v\n", plug)
	pluginend:
	}
	//
	if len(*finalDumpPtr) > 0 {
		time.Sleep(10 * time.Millisecond)
		log.Printf("monimelt should final dump in %s\n", *finalDumpPtr)
		objvalmo.DumpIntoDirectory(*finalDumpPtr)
		log.Printf("monimelt did final dump in %s\n", *finalDumpPtr)
	}
	log.Printf("Monimelt ending pid %d\n", os.Getpid())
}
