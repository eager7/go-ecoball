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

type Chain struct {
	Hash    common.Hash
	TxHash  common.Hash
	Address common.Address
	Index   common.AccountName
}
type Resource struct {
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
	log.Debug("SetResourceLimits:", from, to, cpuStaked, netStaked, cpuStakedSum, netStakedSum)
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
		if err := s.CommitAccount(accTo); err != nil {
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
	return s.CommitAccount(acc)
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
	return s.CommitAccount(acc)
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
		if err := s.CommitAccount(accTo); err != nil {
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
	return s.CommitAccount(acc)
}

/**
 *  @brief recover a account's resource by time
 *  @param index - account's index
 *  @param timeStamp - current time
 */
func (s *State) RecoverResources(index common.AccountName, timeStamp int64, cpuLimit, netLimit float64) error {
	log.Debug("recover resource:", timeStamp)
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
	return s.CommitAccount(acc)
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
	nAcc, err := acc.Clone()
	if err != nil {
		return 0, 0, err
	}
	nAcc.RecoverResources(cpuStakedSum, netStakedSum, timeStamp, cpuLimit, netLimit)
	log.Debug("cpu:", nAcc.Cpu.Used, nAcc.Cpu.Available, nAcc.Cpu.Limit)
	log.Debug("net:", nAcc.Net.Used, nAcc.Net.Available, nAcc.Net.Limit)
	return nAcc.Cpu.Available, nAcc.Net.Available, nil
}

/**
 *  @brief register a new transaction chain
 *  @param index - account's index
 */
func (s *State) RegisterChain(index common.AccountName, hash, txHash common.Hash, addr common.Address) error {
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
	s.Chains[hash] = Chain{
		Hash:    hash,
		TxHash:  txHash,
		Address: addr,
		Index:   index,
	}
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
		List = append(List, s.Chains[hash])
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
				s.Chains[v.Hash] = v
			}
		}
	}
	var list []Chain
	for _, v := range s.Chains {
		c := Chain{
			Hash:    v.Hash,
			TxHash:  v.TxHash,
			Address: v.Address,
			Index:   v.Index,
		}
		list = append(list, c)
		log.Debug(c.Hash.HexString(), c.Index.String())
	}
	return list, nil
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
func (a *Account) CancelOthersDelegate(cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64, cpuLimit, netLimit float64) {
	a.Cpu.Delegated -= cpuStaked
	a.Net.Delegated -= netStaked
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
			acc.CancelOthersDelegate(cpuStaked, netStaked, cpuStakedSum, netStakedSum, cpuLimit, netLimit)

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
		log.Warn(a.JsonString())
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
	for i, d := range a.Delegates {
		if d.Index == index {
			a.Delegates[i].CpuStaked += cpuStaked
			a.Delegates[i].NetStaked += netStaked
			return
		}
	}
	d := Delegate{Index: index, CpuStaked: cpuStaked, NetStaked: netStaked}
	a.Delegates = append(a.Delegates, d)
}
func (a *Account) updateResource(cpuStakedSum, netStakedSum uint64, cpuLimit, netLimit float64) {
	a.Cpu.Limit = float64(a.Cpu.Staked+a.Cpu.Delegated) / float64(cpuStakedSum) * cpuLimit
	a.Cpu.Available = a.Cpu.Limit - a.Cpu.Used
	a.Net.Limit = float64(a.Net.Staked+a.Net.Delegated) / float64(netStakedSum) * netLimit
	a.Net.Available = a.Net.Limit - a.Net.Used
}
func (a *Account) RecoverResources(cpuStakedSum, netStakedSum uint64, timeStamp int64, cpuLimit, netLimit float64) error {
	if timeStamp < a.TimeStamp {
		log.Warn("the transaction could be earlier deal")
		return nil
	}
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
