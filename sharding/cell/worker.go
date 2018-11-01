package cell

import (
	"container/list"
	"github.com/ecoball/go-ecoball/common"
	sc "github.com/ecoball/go-ecoball/sharding/common"
)

type workerSet struct {
	max    int
	member []*sc.Worker
}

func makeWorkerSet(max int) *workerSet {
	return &workerSet{
		max:    max,
		member: make([]*sc.Worker, 0, max),
	}
}

func (s *workerSet) addMember(w *sc.Worker) *sc.Worker {
	length := len(s.member)
	if length == 0 {
		s.member = append(s.member, w)
		return nil
	} else if length == s.max {
		result := make([]*sc.Worker, 0, s.max)
		result = append(result, w)
		result = append(result, s.member[:length-1]...)
		tail := s.member[length-1]

		s.member = result
		return tail
	} else if length < s.max {
		result := make([]*sc.Worker, 0, s.max)
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
	result := make([]*sc.Worker, 0, s.max)
	result = append(result, s.member[1:]...)
	result = append(result, leader)

	s.member = result
}

func (s *workerSet) isLeader(self *sc.Worker) bool {
	if len(s.member) <= 0 {
		return false
	}

	if self.Equal(s.member[0]) {
		return true
	} else {
		return false
	}
}

func (s *workerSet) isBackup(self *sc.Worker) bool {
	if len(s.member) <= 1 {
		return false
	}

	if self.Equal(s.member[1]) {
		return true
	} else {
		return false
	}
}

func (s *workerSet) isMember(self *sc.Worker) bool {
	for _, work := range s.member {
		if work.Equal(self) {
			return true
		}
	}

	return false
}

func (s *workerSet) changeLeader(leader *sc.Worker) {
	for i, work := range s.member {
		if work.Equal(leader) {
			if i == 0 {
				log.Debug("leader not change")
				return
			}

			result := make([]*sc.Worker, 0, s.max)
			result = append(result, s.member[i:]...)
			for j := i - 1; j >= 0; j-- {
				result = append(result, s.member[j])
			}

			log.Debug("new leader i ", i)
			s.member = result
			return
		}
	}

	log.Error("new leader not in committee")
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

func (c *workerQ) addMember(w *sc.Worker) *sc.Worker {
	if c.member.Length() < c.max {
		c.member.Push(w)
		return nil
	} else {
		pop := c.member.Pop()
		c.member.Push(w)
		return pop.(*sc.Worker)
	}
}

func (c *workerQ) isLeader(self *sc.Worker) bool {
	head := c.member.GetHeadValue()
	if head == nil {
		log.Error("cm member is empty")
		return false
	}

	if self.Equal(head.(*sc.Worker)) {
		return true
	} else {
		return false
	}
}

func (c *workerQ) isCandidateLeader(self *sc.Worker) bool {
	head := c.member.GetHead()
	if head == nil {
		log.Error("cm member is empty")
		return false
	}

	next := head.(*list.Element).Next()
	if next == nil {
		return false
	}

	if self.Equal(next.Value.(*sc.Worker)) {
		return true
	} else {
		return false
	}
}
