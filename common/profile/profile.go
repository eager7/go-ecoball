package profile

import (
	"github.com/ecoball/go-ecoball/common/config"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"fmt"
	"runtime/trace"
)

var fc *os.File

func CpuProfile() {
	var err error
	fc, err = os.Create(config.LogDir + "ProfileCpu")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(fc); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
}

func MemProfile() {
	fm, err := os.Create(config.LogDir + "ProfileMem")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(fm); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	fm.Close()
}

func TraceProfile() {
	f, err := os.Create(config.LogDir + "ProfileTrace")
	if err != nil {
		log.Fatal("could not create trace profile: ", err)
	}
	defer f.Close()

	log.Println("Trace started")
	trace.Start(f)
	defer trace.Stop()
}

func StopProfile() {
	pprof.StopCPUProfile()
	fc.Close()
	fmt.Println("complete pprof collect")
}
