package cell

import (
	"container/list"
	"github.com/ecoball/go-ecoball/common"
	cs "github.com/ecoball/go-ecoball/core/shard"
	"github.com/ecoball/go-ecoball/sharding/simulate"
)

type Worker struct {
	Pubkey  string
	Address string
	Port    string
}

func (a *Worker) Equal(b *Worker) bool {
	return a.Pubkey == b.Pubkey
}

func (a *Worker) EqualNode(b *cs.NodeInfo) bool {
	bkey := string(b.PublicKey)
	return a.Pubkey == bkey
}

func (a *Worker) Copy(b *simulate.NodeConfig) {
	a.Pubkey = b.Pubkey
	a.Address = b.Address
	a.Port = b.Port
}

func (a *Worker) InitWork(b *cs.NodeInfo) {
	a.Pubkey = string(b.PublicKey)
	a.Address = b.Address
	a.Port = b.Port
}

type workerSet struct {
	max    int
	member []*Worker
}

func makeWorkerSet(max int) *workerSet {
	return &workerSet{
		max:    max,
		member: make([]*Worker, 0, max),
	}
}

func (s *workerSet) addMember(w *Worker) *Worker {
	length := len(s.member)
	if length == 0 {
		s.member = append(s.member, w)
		return nil
	} else if length == s.max {
		result := make([]*Worker, 0, s.max)
		result = append(result, w)
		result = append(result, s.member[:length-1]...)
		tail := s.member[length-1]

		s.member = result
		return tail
	} else if length < s.max {
		result := make([]*Worker, 0, s.max)
		result = append(result, w)
		result = append(result, s.member[:length]...)

		s.member = result
		return nil
	} else {
		panic("wrong set len")
	}
}

func (s *workerSet) popLeader() {
	leader := s.member[0]
	result := make([]*Worker, 0, s.max)
	result = append(result, s.member[1:]...)
	result = append(result, leader)

	s.member = result
}

func (s *workerSet) isLeader(self *Worker) bool {
	if len(s.member) <= 0 {
		return false
	}

	if self.Equal(s.member[0]) {
		return true
	} else {
		return false
	}
}

func (s *workerSet) isBackup(self *Worker) bool {
	if len(s.member) <= 1 {
		return false
	}

	if self.Equal(s.member[1]) {
		return true
	} else {
		return false
	}
}

func (s *workerSet) isMember(self *Worker) bool {
	for _, work := range s.member {
		if work.Equal(self) {
			return true
		}
	}

	return false
}

func (s *workerSet) resetNewLeader(leader *Worker) {
	for i, work := range s.member {
		if work.Equal(leader) {
			if i == 0 {
				return
			}

			result := make([]*Worker, 0, s.max)
			result = append(result, s.member[i:]...)
			for j := i - 1; j >= 0; j-- {
				result = append(result, s.member[j])
			}

			s.member = result
		}
	}
}

type workerQ struct {
	max    int
	member *common.Queue
}

func makeworkerQ(max int) *workerQ {
	return &workerQ{
		max:    max,
		member: common.NewQueue(),
	}
}

func (c *workerQ) addMember(w *Worker) *Worker {
	if c.member.Length() < c.max {
		c.member.Push(w)
		return nil
	} else {
		pop := c.member.Pop()
		c.member.Push(w)
		return pop.(*Worker)
	}
}

func (c *workerQ) isLeader(self *Worker) bool {
	head := c.member.GetHeadValue()
	if head == nil {
		log.Error("cm member is empty")
		return false
	}

	if self.Equal(head.(*Worker)) {
		return true
	} else {
		return false
	}
}

func (c *workerQ) isCandidateLeader(self *Worker) bool {
	head := c.member.GetHead()
	if head == nil {
		log.Error("cm member is empty")
		return false
	}

	next := head.(*list.Element).Next()
	if next == nil {
		return false
	}

	if self.Equal(next.Value.(*Worker)) {
		return true
	} else {
		return false
	}
}
