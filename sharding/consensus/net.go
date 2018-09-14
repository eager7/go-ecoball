package consensus

import (
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"math"
)

func (c *Consensus) sendToLeader(packet *sc.CsPacket) {
	work := c.ns.GetLeader()
	if work == nil {
		log.Error("leader is nil")
		return
	}

	simulate.Sendto(work.Address, work.Port, packet)
}

func (c *Consensus) GossipBlock(csp *sc.CsPacket) {
	works := c.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	total := int(len(works))
	if total < 5 {
		return
	}

	peers := works[1:]
	size := total - 1
	number := int(math.Sqrt(float64(size)))

	arr := make([]int, 0, 3*size)
	for i := 0; i < 3; i++ {
		for j := 0; j < size; j++ {
			arr = append(arr, j)
		}
	}

	var indexs []int
	for i, peer := range peers {
		if c.ns.Self.Equal(peer) {
			indexs = append(indexs, arr[i+size-number/2:i+size]...)
			indexs = append(indexs, arr[i+size+1:i+size-number/2+number]...)
			break
		}
	}

	//
	//rand.Seed(time.Now().UnixNano())
	//var indexs []int32
	//for i := 0; i < number; i++ {
	//	r := rand.Int31n(total)
	//	if r == 0 {
	//		continue
	//	}
	//
	//	same := false
	//	for _, index := range indexs {
	//		if index == r {
	//			same = true
	//			break
	//		}
	//	}
	//
	//	if same {
	//		continue
	//	}
	//
	//	indexs = append(indexs, r)
	//}

	for _, index := range indexs {
		if c.ns.Self.Equal(works[index]) {
			continue
		}

		log.Debug("gossip to peer address ", works[index].Address, " port ", works[index].Port)
		go simulate.Sendto(works[index].Address, works[index].Port, csp)
	}
}

func (c *Consensus) BroadcastBlock(packet *sc.CsPacket) {
	works := c.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	for _, work := range works {
		if c.ns.Self.Equal(work) {
			continue
		}

		go simulate.Sendto(work.Address, work.Port, packet)
	}
}
