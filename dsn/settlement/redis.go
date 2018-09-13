package settlement

import (
	"fmt"
)

func (s *Settler)getEcoballTotalCap() uint64 {
	rKey := "host_*"
	kHosts := s.rClient.Keys(rKey)
	var tSize uint64
	for _, v := range kHosts.Val() {
		ret := s.rClient.HGet(v, "total")
		if ret.Err() != nil {
			continue
		}
		size, err := ret.Int64()
		if err != nil {
			tSize = tSize + uint64(size)
		}
	}
	return tSize
}

func (s *Settler) getHostTotalCap(pk string) uint64 {
	pKey := fmt.Sprintf("host_%s", pk)
	ret := s.rClient.HGet(pKey, "total")
	if ret.Err() != nil {
		return 0
	}
	size, _ := ret.Int64()
	return uint64(size)
}

func (s *Settler)getEcoballRepoSize() uint64 {
	//TODO
	return 0
}

func (s *Settler)getHostRepoSize(pk string) uint64 {
	//TODO
	return 0
}

func (s *Settler)getHostOnlineTime(pk string) uint32 {
	//TODO
	return 0
}

func (s *Settler)getRenterUsedSize(pk string) uint64 {
	//TODO
	return 0
}
