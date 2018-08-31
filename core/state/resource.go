package state

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"math/big"
	"sort"
)

var cpuAmount = "cpu_amount"
var netAmount = "net_amount"
var prodsList = "prods_list"
var chainList = "chain_list"
var votingAmount = "voting_amount"
var flag = false

const VotesLimit = 200
const ChainLimit = 200

//var BlockCpu = BlockCpuLimit
//var BlockNet = BlockNetLimit
type Producer struct {
	Index  common.AccountName
	Amount uint64
}
type Chain struct {
	Hash common.Hash
	Index common.AccountName
}
type Resource struct {
	Ram struct {
		Quota float64 `json:"quota"`
		Used  float64 `json:"used"`
	}
	Net struct {
		Staked    uint64  `json:"staked_aba, omitempty"`     //total stake delegated from account to self, uint ABA
		Delegated uint64  `json:"delegated_aba, omitempty"`  //total stake delegated to account from others, uint ABA
		Used      float64 `json:"used_byte, omitempty"`      //uint Byte
		Available float64 `json:"available_byte, omitempty"` //uint Byte
		Limit     float64 `json:"limit_byte, omitempty"`     //uint Byte
	}
	Cpu struct {
	Staked    uint64  `json:"staked_aba, omitempty"`    //total stake delegated from account to self, uint ABA
	Delegated uint64  `json:"delegated_aba, omitempty"` //total stake delegated to account from others, uint ABA
	Used      float64 `json:"used_ms, omitempty"`       //uint ms
	Available float64 `json:"available_ms, omitempty"`  //uint ms
	Limit     float64 `json:"limit_ms, omitempty"`      //uint ms
}
	Votes struct {
		Staked    uint64                        `json:"staked_aba, omitempty"` //total stake delegated, uint ABA
		Producers map[common.AccountName]uint64 `json:"producers, omitempty"`  //support nodes' list
	}
}

type Delegate struct {
	Index     common.AccountName `json:"index"`
	CpuStaked uint64             `json:"cpu_aba"`
	NetStaked uint64             `json:"net_aba"`
}

type BlockLimit struct {
	VirtualBlockCpuLimit uint64
	VirtualBlockNetLimit uint64
	BlockCpuLimit        uint64
	BlockNetLimit        uint64
}

/**
 *  @brief set the cpu and net resource to account
 *  @param from - the account which spend aba token
 *  @param to - the account which receive delegated resource
 *  @param cpuStaked - stake delegated cpu
 *  @param netStaked - stake delegated net
 */
func (s *State) SetResourceLimits(from, to common.AccountName, cpuStaked, netStaked uint64, cpuLimit, netLimit float64) error {
	cpuStakedSum, err := s.getParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.getParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	if from == to {
		acc.AddResourceLimits(true, cpuStaked, netStaked, cpuStaked+cpuStakedSum, netStaked+netStakedSum, cpuLimit, netLimit)
	} else {
		acc.SetDelegateInfo(to, cpuStaked, netStaked)
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		accTo.mutex.Lock()
		defer accTo.mutex.Unlock()
		accTo.AddResourceLimits(false, cpuStaked, netStaked, cpuStaked+cpuStakedSum, netStaked+netStakedSum, cpuLimit, netLimit)
		if err := s.commitAccount(accTo); err != nil {
			return err
		}
	}

	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpuStaked)), new(big.Int).SetUint64(uint64(netStaked)))
	if err := acc.SubBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.commitParam(cpuAmount, cpuStaked+cpuStakedSum); err != nil {
		return err
	}
	if err := s.commitParam(netAmount, netStaked+netStakedSum); err != nil {
		return err
	}
	acc.addVotes(cpuStaked + netStaked)
	if err := s.updateElectedProducers(acc, acc.Votes.Staked-cpuStaked-netStaked); err != nil {
		return err
	}
	return s.commitAccount(acc)
}

/**
 *  @brief sub resource from a account
 *  @param index - the account which sub delegated resource
 *  @param cpu - the amount of cpu spend
 *  @param net - the amount of net spend
 */
func (s *State) SubResources(index common.AccountName, cpu, net float64, cpuLimit, netLimit float64) error {
	cpuStakedSum, err := s.getParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.getParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	if err := acc.SubResourceLimits(cpu, net, cpuStakedSum, netStakedSum, cpuLimit, netLimit); err != nil {
		return err
	}
	return s.commitAccount(acc)
}

/**
 *  @brief recycle token, this action was initiated voluntarily
 *  @param from - the account which recycle aba token
 *  @param to - the account which hold aba token
 *  @param cpuStaked - stake delegated cpu
 *  @param netStaked - stake delegated net
 */
func (s *State) CancelDelegate(from, to common.AccountName, cpuStaked, netStaked uint64, cpuLimit, netLimit float64) error {
	votingSum, err := s.getParam(votingAmount)
	if err != nil {
		return err
	}
	cpuStakedSum, err := s.getParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.getParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()

	if from != to {
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		accTo.mutex.Lock()
		defer accTo.mutex.Unlock()
		if err := acc.CancelDelegateOther(accTo, cpuStaked, netStaked, cpuStakedSum, netStakedSum, cpuLimit, netLimit); err != nil {
			return err
		}
		if err := s.commitAccount(accTo); err != nil {
			return err
		}
	} else {
		acc.CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum, cpuLimit, netLimit)
	}
	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpuStaked)), new(big.Int).SetUint64(uint64(netStaked)))
	if err := acc.AddBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.commitParam(cpuAmount, cpuStakedSum-cpuStaked); err != nil {
		return err
	}
	if err := s.commitParam(netAmount, netStakedSum-cpuStaked); err != nil {
		return err
	}
	if err := s.commitParam(votingAmount, votingSum-cpuStaked-netStaked); err != nil {
		return err
	}
	valueOld := acc.Resource.Votes.Staked
	acc.subVotes(cpuStaked + netStaked)
	if err := s.updateElectedProducers(acc, valueOld); err != nil {
		return err
	}
	if acc.Votes.Staked < VotesLimit {
		s.prodMutex.Lock()
		delete(s.Producers, acc.Index)
		s.prodMutex.Unlock()
	}
	s.commitProducersList()
	return s.commitAccount(acc)
}

/**
 *  @brief recover a account's resource by time
 *  @param index - account's index
 *  @param timeStamp - current time
 */
func (s *State) RecoverResources(index common.AccountName, timeStamp int64, cpuLimit, netLimit float64) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	cpuStakedSum, err := s.getParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.getParam(netAmount)
	if err != nil {
		return err
	}
	acc.RecoverResources(cpuStakedSum, netStakedSum, timeStamp, cpuLimit, netLimit)
	return s.commitAccount(acc)
}

/**
 *  @brief require a account's resource info
 *  @param index - account's index
 *  @param timeStamp - current time
 */
func (s *State) RequireResources(index common.AccountName, cpuLimit, netLimit float64, timeStamp int64) (float64, float64, error) {
	cpuStakedSum, err := s.getParam(cpuAmount)
	if err != nil {
		return 0, 0, err
	}
	netStakedSum, err := s.getParam(netAmount)
	if err != nil {
		return 0, 0, err
	}
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return 0, 0, err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	acc.RecoverResources(cpuStakedSum, netStakedSum, timeStamp, cpuLimit, netLimit)
	log.Debug("cpu:", acc.Cpu.Used, acc.Cpu.Available, acc.Cpu.Limit)
	log.Debug("net:", acc.Net.Used, acc.Net.Available, acc.Net.Limit)
	return acc.Cpu.Available, acc.Net.Available, nil
}

/**
 *  @brief register as a candidate node
 *  @param index - account's index
 */
func (s *State) RegisterProducer(index common.AccountName) error {
	if err := s.initProducersList(); err != nil {
		return err
	}
	s.prodMutex.Lock()
	if _, ok := s.Producers[index]; ok {
		s.prodMutex.Unlock()
		return errors.New(log, fmt.Sprintf("the account:%s was already registed", index.String()))
	}
	if err := s.checkAccountCertification(index, VotesLimit); err != nil {
		s.prodMutex.Unlock()
		return nil
	}
	s.Producers[index] = 0
	s.prodMutex.Unlock()
	return s.commitProducersList()
}

/**
 *  @brief register a new transaction chain
 *  @param index - account's index
 */
func (s *State) RegisterChain(index common.AccountName, hash common.Hash) error {
	if _, err := s.GetChainList(); err != nil {
		return err
	}
	s.chainMutex.Lock()
	if _, ok := s.Chains[hash]; ok {
		s.chainMutex.Unlock()
		return errors.New(log, fmt.Sprintf("the chain:%s was already registed", hash.HexString()))
	}
	if err := s.checkAccountCertification(index, ChainLimit); err != nil {
		s.chainMutex.Unlock()
		return nil
	}
	s.Chains[hash] = index
	s.chainMutex.Unlock()

	return s.commitChains()
}
func (s *State) commitChains() error {
	if len(s.Chains) == 0 {
		return nil
	}

	s.chainMutex.Lock()
	defer s.chainMutex.Unlock()
	var Keys []string
	for k := range s.Chains {
		Keys = append(Keys, k.HexString())
	}
	sort.Strings(Keys)
	var List []Chain
	for _, v := range Keys {
		hash := common.HexToHash(v)
		list := Chain{hash, s.Chains[hash]}
		List = append(List, list)
	}

	data, err := json.Marshal(List)
	if err != nil {
		return errors.New(log, fmt.Sprintf("error convert to json string:%s", err.Error()))
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.trie.TryUpdate([]byte(chainList), data); err != nil {
		return errors.New(log, fmt.Sprintf("error update trie:%s", err.Error()))
	}
	return nil
}
func (s *State) GetChainList() ([]Chain, error) {
	s.chainMutex.Lock()
	defer s.chainMutex.Unlock()
	if len(s.Chains) == 0 {
		data, err := s.trie.TryGet([]byte(chainList))
		if err != nil {
			return nil, errors.New(log, fmt.Sprintf("can't get chainList from DB:%s", err.Error()))
		}
		if len(data) != 0 {
			var Chains []Chain
			if err := json.Unmarshal(data, &Chains); err != nil {
				return nil, errors.New(log, fmt.Sprintf("can't unmarshal Chains List from json string:%s", err.Error()))
			}
			for _, v := range Chains {
				s.Chains[v.Hash] = v.Index
			}
		}
	}
	var list []Chain
	for k := range s.Chains {
		c := Chain{k, s.Chains[k]}
		list = append(list, c)
	}
	return list, nil
}
/**
 *  @brief cancel register as a candidate node
 *  @param index - account's index
 */
func (s *State) UnRegisterProducer(index common.AccountName) error {
	if err := s.initProducersList(); err != nil {
		return err
	}
	s.prodMutex.Lock()
	if _, ok := s.Producers[index]; !ok {
		s.prodMutex.Unlock()
		return errors.New(log, fmt.Sprintf("the account:%s is not registed", index.String()))
	} else {
		delete(s.Producers, index)
	}
	s.prodMutex.Unlock()
	return s.commitProducersList()
}

/**
 *  @brief vote to candidate nodes
 *  @param index - account index
 *  @param accounts - candidate node list
 */
func (s *State) ElectionToVote(index common.AccountName, accounts []common.AccountName) error {
	votingSum, err := s.getParam(votingAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	if acc.Resource.Votes.Staked == 0 {
		return errors.New(log, fmt.Sprintf("the account:%s has no enough vote", index.String()))
	}
	if err := s.initProducersList(); err != nil {
		return err
	}
	s.prodMutex.RLock()
	for _, v := range accounts {
		if _, ok := s.Producers[v]; !ok {
			s.prodMutex.Unlock()
			return errors.New(log, fmt.Sprintf("the account:%s is not register", v.String()))
		}
	}
	s.prodMutex.RUnlock()
	if err := s.changeElectedProducers(acc, accounts); err != nil {
		return err
	}
	if err := s.commitParam(votingAmount, votingSum+acc.Resource.Votes.Staked); err != nil {
		return err
	}
	if votingSum+acc.Resource.Votes.Staked > abaTotal*0.15 && flag == false {
		flag = true
		log.Warn("Start Process ##################################################################################")
		producers, err := s.GetProducerList()
		if err != nil {
			return err
		}
		var accFactors []AccFactor
		for _, v := range producers {
			accFactor := AccFactor{Actor: v, Weight: 1, Permission: Active}
			accFactors = append(accFactors, accFactor)
		}
		perm := NewPermission(Active, Owner, 2, []KeyFactor{}, accFactors)
		root, err := s.GetAccountByName(common.NameToIndex("root"))
		if err != nil {
			return err
		}
		root.mutex.Lock()
		defer root.mutex.Unlock()
		root.AddPermission(perm)
		//if config.ConsensusAlgorithm != "SOLO" {
		//	log.Info(event.Send(event.ActorNil, event.ActorConsensusSolo, &message.SoloStop{}))
		//	log.Info(event.Send(event.ActorNil, event.ActorConsensus, &message.ABABFTStart{}))
		//}
	}
	return s.commitAccount(acc)
}

/**
 *  @brief revote to another nodes
 *  @param acc - account struct
 *  @param accounts - candidate node list
 */
func (s *State) changeElectedProducers(acc *Account, accounts []common.AccountName) error {
	if err := s.initProducersList(); err != nil {
		return err
	}
	s.prodMutex.Lock()
	for k := range acc.Votes.Producers {
		if _, ok := s.Producers[k]; ok {
			s.Producers[k] = s.Producers[k] - acc.Votes.Producers[k]
		}
		delete(acc.Votes.Producers, k)
	}
	for _, v := range accounts {
		if err := s.checkAccountCertification(v, VotesLimit); err != nil {
			s.prodMutex.Unlock()
			return err
		}
		acc.Votes.Producers[v] = acc.Votes.Staked
		if _, ok := s.Producers[v]; !ok {
			s.prodMutex.Unlock()
			return errors.New(log, fmt.Sprintf("the account:%s is not a candidata node", v.String()))
		}
		s.Producers[v] += acc.Votes.Staked
	}
	s.prodMutex.Unlock()
	return s.commitProducersList()
}

/**
 *  @brief update producers when the account change cpu or net resource
 *  @param acc - account struct
 *  @param votesOld - account's votes before changed
 */
func (s *State) updateElectedProducers(acc *Account, votesOld uint64) error {
	if err := s.initProducersList(); err != nil {
		return err
	}
	s.prodMutex.Lock()
	for k := range acc.Votes.Producers {
		acc.Votes.Producers[k] = acc.Votes.Staked
		if _, ok := s.Producers[k]; ok {
			s.Producers[k] = s.Producers[k] - votesOld + acc.Votes.Staked
		} else {
			s.prodMutex.Unlock()
			return errors.New(log, fmt.Sprintf("the account:%s is exit candidata nodes list", k.String()))
		}
	}
	s.prodMutex.Unlock()
	return s.commitProducersList()
}

/**
 *  @brief check whether the account is qualified
 *  @param index - account's index
 */
func (s *State) checkAccountCertification(index common.AccountName, votes uint64) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if acc.Votes.Staked < votes {
		acc.Show()
		return errors.New(log, fmt.Sprintf("the account:%s has no enough staked:%d", index.String(), acc.Votes.Staked))
	}
	return nil
}

/**
 *  @brief store the producers' list into mpt trie
 */
func (s *State) commitProducersList() error {
	//if err := s.initProducersList(); err != nil {
	//	return err
	//}
	s.prodMutex.Lock()
	defer s.prodMutex.Unlock()
	var Keys []common.AccountName
	for k := range s.Producers {
		Keys = append(Keys, k)
	}
	sort.Slice(Keys, func(i, j int) bool {
		return uint64(Keys[i]) > uint64(Keys[j])
	})
	var List []Producer
	for _, v := range Keys {
		list := Producer{v, s.Producers[v]}
		List = append(List, list)
	}

	data, err := json.Marshal(List)
	if err != nil {
		return errors.New(log, fmt.Sprintf("error convert to json string:%s", err.Error()))
	}
	if err := s.trie.TryUpdate([]byte(prodsList), data); err != nil {
		return errors.New(log, fmt.Sprintf("error update trie:%s", err.Error()))
	}
	return nil
}

/**
 *  @brief require the voting information
 */
func (s *State) RequireVotingInfo() bool {
	votingSum, err := s.getParam(votingAmount)
	if err != nil {
		return false
	}
	log.Debug("abaTotal", abaTotal, "votingSum", votingSum, "Percentage", 100*float32(votingSum)/float32(abaTotal), "%")
	if float32(votingSum)/float32(abaTotal) >= 0.15 {
		return true
	}
	return false
}

func (s *State) GetProducerList() ([]common.AccountName, error) {
	if !s.RequireVotingInfo() {
		return nil, errors.New(log, "the main network has not been started")
	}
	if err := s.initProducersList(); err != nil {
		return nil, err
	}
	var list []common.AccountName
	s.prodMutex.RLock()
	defer s.prodMutex.RUnlock()
	for k := range s.Producers {
		list = append(list, k)
		if len(list) == 21 {
			break
		}
	}
	return list, nil
}

func (s *State) initProducersList() error {
	if len(s.Producers) == 0 {
		s.prodMutex.Lock()
		defer s.prodMutex.Unlock()
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		data, err := s.trie.TryGet([]byte(prodsList))
		if err != nil {
			return errors.New(log, fmt.Sprintf("can't get ProdList from DB:%s", err.Error()))
		}
		if len(data) != 0 {
			var Producers []Producer
			if err := json.Unmarshal(data, &Producers); err != nil {
				return errors.New(log, fmt.Sprintf("can't unmarshal ProdList from json string:%s", err.Error()))
			}
			for _, v := range Producers {
				s.Producers[v.Index] = v.Amount
			}
		}
	}
	return nil
}

/**
 *  @brief set the cpu and net resource to account
 *  @param self - if self, set resource to staked, otherwise, set resource to delegated
 *  @param cpuStaked - stake delegated cpu
 *  @param netStaked - stake delegated net
 *  @param cpuStakedSum - total stake cpu
 *  @param netStakedSum - total stake net
 */
func (a *Account) AddResourceLimits(self bool, cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64, cpuLimit, netLimit float64) {
	if self {
		a.Cpu.Staked += cpuStaked
		a.Net.Staked += netStaked
	} else {
		a.Cpu.Delegated += cpuStaked
		a.Net.Delegated += netStaked
	}
	a.updateResource(cpuStakedSum, netStakedSum, cpuLimit, netLimit)
}
func (a *Account) CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64, cpuLimit, netLimit float64) {
	a.Cpu.Staked -= cpuStaked
	a.Net.Staked -= netStaked
	a.updateResource(cpuStakedSum, netStakedSum, cpuLimit, netLimit)
}
func (a *Account) CancelDelegateOther(acc *Account, cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64, cpuLimit, netLimit float64) error {
	done := false
	for i := 0; i < len(a.Delegates); i++ {
		if a.Delegates[i].Index == acc.Index {
			done = true
			if acc.Cpu.Delegated < cpuStaked {
				return errors.New(log, fmt.Sprintf("the account:%s cpu amount is not enough", acc.Index.String()))
			}
			if acc.Net.Delegated < netStaked {
				return errors.New(log, fmt.Sprintf("the account:%s net amount is not enough", acc.Index.String()))
			}
			acc.CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum, cpuLimit, netLimit)

			a.Delegates[i].CpuStaked -= cpuStaked
			a.Delegates[i].NetStaked -= netStaked
			if a.Delegates[i].CpuStaked == 0 && a.Delegates[i].NetStaked == 0 {
				a.Delegates = append(a.Delegates[:i], a.Delegates[i+1:]...)
			}
		}
	}
	if done == false {
		return errors.New(log, fmt.Sprintf("account:%s is not delegated for %s", a.Index.String(), acc.Index.String()))
	}
	return nil
}
func (a *Account) SubResourceLimits(cpu, net float64, cpuStakedSum, netStakedSum uint64, cpuLimit, netLimit float64) error {
	if a.Cpu.Available < cpu {
		log.Warn(a.JsonString(false))
		return errors.New(log, fmt.Sprintf("the account:%s cpu avaiable[%f] is not enough", a.Index.String(), a.Cpu.Available))
	}
	if a.Net.Available < net {
		return errors.New(log, fmt.Sprintf("the account:%s net avaiable[%f] is not enough", a.Index.String(), a.Net.Available))
	}
	a.Cpu.Used += cpu
	a.Net.Used += net
	a.updateResource(cpuStakedSum, netStakedSum, cpuLimit, netLimit)
	return nil
}
func (a *Account) SetDelegateInfo(index common.AccountName, cpuStaked, netStaked uint64) {
	d := Delegate{Index: index, CpuStaked: cpuStaked, NetStaked: netStaked}
	a.Delegates = append(a.Delegates, d)
}
func (a *Account) updateResource(cpuStakedSum, netStakedSum uint64, cpuLimit, netLimit float64) {
	a.Cpu.Limit = float64(a.Cpu.Staked+a.Cpu.Delegated) / float64(cpuStakedSum) * cpuLimit
	a.Cpu.Available = a.Cpu.Limit - a.Cpu.Used
	a.Net.Limit = float64(a.Cpu.Staked+a.Net.Delegated) / float64(netStakedSum) * netLimit
	a.Net.Available = a.Net.Limit - a.Net.Used
}
func (a *Account) RecoverResources(cpuStakedSum, netStakedSum uint64, timeStamp int64, cpuLimit, netLimit float64) error {
	t := (timeStamp - a.TimeStamp) / (1000 * 1000)
	interval := 100.0 * float64(t) / (24.0 * 60.0 * 60.0 * 1000)
	if interval >= 100 {
		a.Cpu.Used = 0
		a.Net.Used = 0
	}
	if a.Cpu.Used != 0 {
		a.Cpu.Used -= a.Cpu.Used * interval
	}
	if a.Net.Used != 0 {
		a.Net.Used -= a.Net.Used * interval
	}
	a.updateResource(cpuStakedSum, netStakedSum, cpuLimit, netLimit)
	a.TimeStamp = timeStamp
	return nil
}
func (a *Account) addVotes(staked uint64) {
	a.Resource.Votes.Staked += staked
}
func (a *Account) subVotes(staked uint64) {
	a.Resource.Votes.Staked -= staked
}
