package state

import (
	"encoding/json"
	"fmt"
	. "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"sort"
)

//存储选举出来的节点列表
type Producer struct {
	Index  AccountName
	Amount uint64
}

//存储参选节点信息
type Elector struct {
	Index   AccountName
	Amount  uint64
	B64Pub  string
	Address string
	Port    uint32
	Payee   AccountName
}

func (e *Elector) Serialize() ([]byte, error) {
	pbE := pb.Elector{
		Index:   e.Index.Number(),
		Amount:  e.Amount,
		Address: e.Address,
		Port:    e.Port,
		Payee:   e.Payee.Number(),
	}
	data, err := pbE.Marshal()
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return data, err
}

func (e *Elector) Deserialize(data []byte) error {
	pbE := pb.Elector{}
	if err := pbE.Unmarshal(data); err != nil {
		return errors.New(err.Error())
	}
	e.Index = AccountName(pbE.Index)
	e.Amount = pbE.Amount
	e.Address = pbE.Address
	e.Port = pbE.Port
	e.Payee = AccountName(pbE.Payee)
	return nil
}

/**
 *  @brief 取消参选，会将账号从候选列表中删除，但是不会删除其他人为之投票的信息
 *  @param index - account's index
 */
func (s *State) UnRegisterProducer(index AccountName) error {
	if err := s.initProducersList(); err != nil {
		return err
	}
	if producer := s.Producers.Get(index); producer == nil {
		return errors.New(fmt.Sprintf("the account:%s is not registed", index.String()))
	} else {
		s.Producers.Del(index)
	}
	return s.commitProducersList()
}

/**
 *  @brief 投票给候选节点，可以自己给自己投票，但是自己的全部票都要投出，且均等的投给自己想投的节点，票数不能分散，当投票数量大于全网代币15%时，启动主网
 *  @param index - account index
 *  @param accounts - candidate node list
 */
func (s *State) ElectionToVote(index AccountName, accounts []AccountName) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	if acc.Resource.Votes.Staked == 0 {
		return errors.New(fmt.Sprintf("the account:%s has no enough vote", index.String()))
	}
	if err := s.initProducersList(); err != nil {
		return err
	}
	for _, acc := range accounts {
		if producer := s.Producers.Get(acc); producer == nil {
			return errors.New(fmt.Sprintf("the account:%s is not register", acc.String()))
		}
	}
	if err := s.changeElectedProducers(acc, accounts); err != nil {
		return err
	}
	//检查主网是否可以启动
	votingSum, err := s.getParam(votingAmount)
	if err != nil {
		return err
	}
	if err := s.commitParam(votingAmount, votingSum+acc.Resource.Votes.Staked); err != nil {
		return err
	}
	if votingSum+acc.Resource.Votes.Staked > AbaTotal*0.15 && flag == false {
		flag = true
		log.Warn("Start Process ##################################################################################")
		producers, err := s.GetProducerList()
		if err != nil {
			return err
		}
		var accFactors []AccFactor
		for _, v := range producers {
			accFactor := AccFactor{Actor: v.Index, Weight: 1, Permission: Active}
			accFactors = append(accFactors, accFactor)
		}
		perm := NewPermission(Active, Owner, 2, []KeyFactor{}, accFactors)
		root, err := s.GetAccountByName(NameToIndex("root"))
		if err != nil {
			return err
		}
		root.lock.Lock()
		root.AddPermission(perm)
		root.lock.Unlock()
	}
	return s.CommitAccount(acc)
}

/**
 *  @brief 重新分配选票给新的候选人列表，本账户会保存自己支持的候选人，同时MPT会保存候选人票数
 *  @param acc - account struct
 *  @param accounts - candidate node list
 */
func (s *State) changeElectedProducers(acc *Account, accounts []AccountName) error {
	if err := s.initProducersList(); err != nil {
		return err
	}
	for index := range acc.Resource.Votes.Producers { //为防止重复投票，在更新票数前先把之前投的票作废
		if producer := s.Producers.Get(index); producer != nil {
			s.Producers.Add(index, producer.Amount-acc.Resource.Votes.Producers[index])
		}
		delete(acc.Resource.Votes.Producers, index)
	}
	for _, a := range accounts {
		if err := s.checkAccountCertification(a, VotesLimit); err != nil {
			return err
		}
		acc.Resource.Votes.Producers[a] = acc.Resource.Votes.Staked
		if producer := s.Producers.Get(a); producer == nil {
			return errors.New(fmt.Sprintf("the account:%s is not a candidata node", a.String()))
		} else {
			s.Producers.Add(a, producer.Amount+acc.Resource.Votes.Staked)
		}
	}
	return s.commitProducersList()
}

/**
 *  @brief 在用户改变了cpu以及net的抵押量时，他的选票数量会随之改变，此时需要更新他支持的候选人票数
 *  @param acc - account struct
 *  @param votesOld - account's votes before changed
 */
func (s *State) updateElectedProducers(acc *Account, votesOld uint64) error {
	if err := s.initProducersList(); err != nil {
		return err
	}
	for k := range acc.Resource.Votes.Producers {
		acc.Resource.Votes.Producers[k] = acc.Resource.Votes.Staked
		if producer := s.Producers.Get(k); producer != nil {
			s.Producers.Add(k, producer.Amount-votesOld+acc.Resource.Votes.Staked)
		} else {
			return errors.New(fmt.Sprintf("the account:%s is exit candidata nodes list", k.String()))
		}
	}
	return s.commitProducersList()
}

/**
 *  @brief 检查账户是否具有参选资格，需要至少抵押200个代币才能具备参选资格
 *  @param index - account's index
 */
func (s *State) checkAccountCertification(index AccountName, votes uint64) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if acc.Resource.Votes.Staked < votes {
		return errors.New(fmt.Sprintf("the account:%s has no enough staked:%d", index.String(), acc.Resource.Votes.Staked))
	}
	return nil
}

/**
 *  @brief 将候选人列表存到levelDB中，以备程序重启时可以重新获取数据
 */
func (s *State) commitProducersList() error {
	var Keys []AccountName
	for producer := range s.Producers.Iterator() {
		Keys = append(Keys, producer.Index)
	}
	sort.Slice(Keys, func(i, j int) bool {
		return uint64(Keys[i]) > uint64(Keys[j])
	})
	var List []*Producer
	for _, v := range Keys {
		List = append(List, s.Producers.Get(v))
	}

	data, err := json.Marshal(List)
	if err != nil {
		return errors.New(fmt.Sprintf("error convert to json string:%s", err.Error()))
	}
	if err := s.trie.TryUpdate([]byte(prodsList), data); err != nil {
		return errors.New(fmt.Sprintf("error update trie:%s", err.Error()))
	}
	return nil
}

/**
 *  @brief 查询投票情况，如果投票数量大于全网代币15%，表示可以启动主网，返回true
 */
func (s *State) RequireVotingInfo() bool {
	votingSum, err := s.getParam(votingAmount)
	if err != nil {
		return false
	}
	log.Debug("abaTotal", AbaTotal, "votingSum", votingSum, "Percentage", 100*float32(votingSum)/float32(AbaTotal), "%")
	if float32(votingSum)/float32(AbaTotal) >= 0.15 {
		return true
	}
	return false
}

/**
 *  @brief 将选举节点信息返回，返回内容包括候选人账户名，票数，地址以及付款账户
 */
func (s *State) GetProducerList() ([]Elector, error) {
	if err := s.initProducersList(); err != nil {
		return nil, err
	}
	var electors []Elector
	for producer := range s.Producers.Iterator() {
		acc, err := s.GetAccountByName(producer.Index)
		if err != nil {
			return nil, err
		}
		acc.lock.RLock()
		electors = append(electors, acc.Elector)
		acc.lock.RUnlock()
	}
	return electors, nil
}

/**
 *  @brief 在程序刚启动时，从数据库中读取参选节点列表，恢复到Producers映射中，映射里只保存账号名和票数，其余信息需要从对应账号获取
 */
func (s *State) initProducersList() error {
	if s.Producers.Len() == 0 {
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		data, err := s.trie.TryGet([]byte(prodsList))
		if err != nil {
			return errors.New(fmt.Sprintf("can't get ProdList from DB:%s", err.Error()))
		}
		if len(data) != 0 {
			var Producers []Producer
			if err := json.Unmarshal(data, &Producers); err != nil {
				return errors.New(fmt.Sprintf("can't unmarshal ProdList from json string:%s", err.Error()))
			}
			for _, v := range Producers {
				s.Producers.Add(v.Index, v.Amount)
			}
		}
	}
	return nil
}

/**
 *  @brief 注册成为一个候选节点，票数为零，需等待其他节点投票给自己
 *  @param index - account's index
 */
func (s *State) RegisterProducer(index AccountName, b64Pub, addr string, port uint32, payee AccountName) error {
	if _, err := s.GetAccountByName(payee); err != nil {
		return err
	}
	if err := s.initProducersList(); err != nil {
		return err
	}
	if producer := s.Producers.Get(index); producer != nil {
		return errors.New(fmt.Sprintf("the account:%s was already registed", index.String()))
	}
	if err := s.checkAccountCertification(index, VotesLimit); err != nil {
		return err
	}
	s.Producers.Add(index, 0)
	if err := s.commitProducersList(); err != nil {
		return err
	}

	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	acc.Elector.Index = index
	acc.Elector.B64Pub = b64Pub
	acc.Elector.Address = addr
	acc.Elector.Port = port
	acc.Elector.Amount = 0
	acc.Elector.Payee = payee
	return s.CommitAccount(acc)
}
