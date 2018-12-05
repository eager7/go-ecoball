package profile

import (
	"github.com/ecoball/go-ecoball/common/config"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func CpuProfile() {
	f, err := os.Create(config.LogDir + "ProfileCpu")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
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

func StopProfile() {
	pprof.StopCPUProfile()
}
