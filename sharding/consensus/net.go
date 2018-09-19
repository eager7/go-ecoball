package consensus

import (
	"github.com/ecoball/go-ecoball/sharding/cell"
	sc "github.com/ecoball/go-ecoball/sharding/common"
	"github.com/ecoball/go-ecoball/sharding/simulate"
	"math"
	"math/rand"
	"time"
)

func (c *Consensus) sendToPeer(packet *sc.NetPacket, worker *cell.Worker) {
	if worker == nil {
		log.Error("leader is nil")
		return
	}

	go simulate.Sendto(worker.Address, worker.Port, packet)
}

func (c *Consensus) GossipBlock(packet *sc.NetPacket) {
	works := c.ns.GetWorks()
	if works == nil {
		log.Error("works is nil")
		return
	}

	if len(works) < 5 {
		if len(works) > 1 {
			rand.Seed(time.Now().UnixNano())
			r := rand.Int31n(int32(len(works)))
			log.Debug("gossip to peer address ", works[r].Address, " port ", works[r].Port)
			go simulate.Sendto(works[r].Address, works[r].Port, packet)
		}
		return
	}

	peers := works[1:]

	size := int(len(peers))
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
		log.Debug("gossip to peer address ", peers[index].Address, " port ", peers[index].Port)
		go simulate.Sendto(peers[index].Address, peers[index].Port, packet)
	}
}

func (c *Consensus) BroadcastBlock(packet *sc.NetPacket) {
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
