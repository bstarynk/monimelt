// file objvalmo/tiny.go

package objvalmo // import "github.com/bstarynk/monimelt/objvalmo"

import (
	"log"
	osexec "os/exec"
)

func DoTinyDump1(tempdir string) {
	log.Printf("DoTinyDump1 starting tempdir=%q...\n", tempdir)
	osexec.Command("rm", "-rvf", tempdir).Run()
	pr_name := Predef_02hL3RuX4x6_6y6PTK9vZs7()
	log.Printf("DoTinyDump1 pr_name=%v (%T)\n", pr_name, pr_name)
	log.Printf("DoTinyDump1 globalnames=%v\n", NamesGlobalVariables())
	ro1 := NewRefobV()
	ro2 := NewRefobV()
	ro3 := NewRefobV()
	ro4 := NewRefobV()
	ro5 := NewRefobV()
	ro6 := NewRefobV()
	log.Printf("DoTinyDump1 ro1=%v (%T);  ro2=%v (%T)\n",
		ro1, ro1, ro2, ro2)
	log.Printf("DoTinyDump1 ro3=%v (%T);  ro4=%v (%T)\n",
		ro3, ro3, ro4, ro4)
	log.Printf("DoTinyDump1 ro5=%v (%T);  ro6=%v (%T)\n",
		ro5, ro5, ro6, ro6)
	log.Printf("DoTinyDump1 ro1=%v ro2=%v ro3=%v\n", ro1, ro2, ro3)
	log.Printf("DoTinyDump1 ro4=%v ro5=%v ro6=%v\n", ro4, ro5, ro6)
	ro1.Obref().UnsyncSetSpaceNum(SpaUser)
	ro2.Obref().UnsyncSetSpaceNum(SpaUser)
	ro3.Obref().UnsyncSetSpaceNum(SpaUser)
	ro4.Obref().UnsyncSetSpaceNum(SpaUser)
	ro5.Obref().UnsyncSetSpaceNum(SpaUser)
	ro6.Obref().UnsyncSetSpaceNum(SpaUser)
	ro1.Obref().UnsyncPutAttr(pr_name, MakeStringV("ro1"))
	ro2.Obref().UnsyncPutAttr(pr_name, MakeStringV("ro2"))
	ro3.Obref().UnsyncPutAttr(pr_name, MakeStringV("ro3"))
	ro4.Obref().UnsyncPutAttr(pr_name, MakeStringV("ro4"))
	ro5.Obref().UnsyncPutAttr(pr_name, MakeStringV("ro5"))
	ro6.Obref().UnsyncPutAttr(pr_name, MakeStringV("ro6"))
	ro1.Obref().UnsyncAddValues(ro1, ro2, ro3, ro4, ro5, ro6)
	ro1.Obref().UnsyncPutAttr(ro2.Obref(), MakeStringV("some string here for ro1"))
	ro2.Obref().UnsyncPutAttr(ro1.Obref(), ro3)
	ro2.Obref().UnsyncPutAttr(ro5.Obref(), MakeIntV(123))
	ro3.Obref().UnsyncPutAttr(ro2.Obref(), MakeIntV(-1))
	ro3.Obref().UnsyncPutAttr(ro5.Obref(), MakeIntV(12345678901234))
	ro3.Obref().UnsyncPutAttr(ro3.Obref(), MakeTupleV(ro1.Obref(), ro2.Obref(), ro1.Obref(), ro3.Obref(), ro5.Obref()))
	ro4.Obref().UnsyncPutAttr(ro2.Obref(), MakeSetV(ro2.Obref(), ro3.Obref(), ro2.Obref(), ro6.Obref()))
	ro6.Obref().UnsyncPutAttr(ro2.Obref(), MakeFloatV(3.14))
	ro6.Obref().UnsyncPutAttr(ro3.Obref(), MakeFloatV(0.0))
	ro6.Obref().UnsyncPutAttr(ro4.Obref(), MakeFloatV(-1.2345678901e75))
	Glob_the_system = ro1.Obref()
	log.Printf("DoTinyDump1 Glob_the_system=%v\n", Glob_the_system)
	log.Printf("DoTinyDump1 ro1=%v (%T)\n.. == %#v\n\n", ro1, ro1, ro1)
	log.Printf("DoTinyDump1 ro2=%v (%T)\n.. == %#v\n\n", ro2, ro2, ro2)
	log.Printf("DoTinyDump1 ro3=%v (%T)\n.. == %#v\n\n", ro3, ro3, ro3)
	log.Printf("DoTinyDump1 ro4=%v (%T)\n.. == %#v\n\n", ro4, ro4, ro4)
	log.Printf("DoTinyDump1 ro5=%v (%T)\n.. == %#v\n\n", ro5, ro5, ro5)
	log.Printf("DoTinyDump1 ro6=%v (%T)\n.. == %#v\n\n", ro6, ro6, ro6)
	log.Printf("DoTinyDump1 should dump globalnames=%v\n", NamesGlobalVariables())
	log.Printf("DoTinyDump1 Glob_the_system is %v\n", Glob_the_system)
	log.Printf("DoTinyDump1 before dump in %s\n", tempdir)
	log.Printf("DoTinyDump1 before DumpIntoDirectory tempdir %s\n", tempdir)
	DumpIntoDirectory(tempdir)
	log.Printf("DoTinyDump1 after DumpIntoDirectory tempdir %s\n", tempdir)
	log.Printf("DoTinyDump1 ending...\n\n")
} // end DoTinyDump1
