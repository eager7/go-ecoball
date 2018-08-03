package state

import (
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"math/big"
)

var cpuAmount = "cpu_amount"
var netAmount = "net_amount"
var votingAmount = "voting_amount"

const VirtualBlockCpuLimit float32 = 200000000.0
const VirtualBlockNetLimit float32 = 1048576000.0
const BlockCpuLimit float32 = 200000.0
const BlockNetLimit float32 = 1048576.0

var BlockCpu = BlockCpuLimit
var BlockNet = BlockNetLimit

type Resource struct {
	Ram struct {
		Quota float32 `json:"quota"`
		Used  float32 `json:"used"`
	}
	Net struct {
		Staked    uint64  `json:"staked_aba, omitempty"`     //total stake delegated from account to self, uint ABA
		Delegated uint64  `json:"delegated_aba, omitempty"`  //total stake delegated to account from others, uint ABA
		Used      float32 `json:"used_byte, omitempty"`      //uint Byte
		Available float32 `json:"available_byte, omitempty"` //uint Byte
		Limit     float32 `json:"limit_byte, omitempty"`     //uint Byte
	}
	Cpu struct {
		Staked    uint64  `json:"staked_aba, omitempty"`    //total stake delegated from account to self, uint ABA
		Delegated uint64  `json:"delegated_aba, omitempty"` //total stake delegated to account from others, uint ABA
		Used      float32 `json:"used_ms, omitempty"`       //uint ms
		Available float32 `json:"available_ms, omitempty"`  //uint ms
		Limit     float32 `json:"limit_ms, omitempty"`      //uint ms
	}
	Votes struct {
		Staked    uint64
		Producers map[common.AccountName]uint64
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
func (s *State) SetResourceLimits(from, to common.AccountName, cpuStaked, netStaked uint64) error {
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	acc.addVotes(cpuStaked + netStaked)
	if from == to {
		if err := acc.AddResourceLimits(true, cpuStaked, netStaked, cpuStaked+cpuStakedSum, netStaked+netStakedSum); err != nil {
			return err
		}
	} else {
		if err := acc.SetDelegateInfo(to, cpuStaked, netStaked); err != nil {
			return err
		}
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		if err := accTo.AddResourceLimits(false, cpuStaked, netStaked, cpuStaked+cpuStakedSum, netStaked+netStakedSum); err != nil {
			return err
		}
		if err := s.CommitAccount(accTo); err != nil {
			return err
		}
	}

	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpuStaked)), new(big.Int).SetUint64(uint64(netStaked)))
	if err := acc.SubBalance(AbaToken, value); err != nil {
		return err
	}
	if err := s.CommitParam(cpuAmount, cpuStaked+cpuStakedSum); err != nil {
		return err
	}
	if err := s.CommitParam(netAmount, netStaked+netStakedSum); err != nil {
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
func (s *State) SubResources(index common.AccountName, cpu, net float32) error {
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	//acc.RecoverResources(cpuStakedSum, netStakedSum)
	if err := acc.SubResourceLimits(cpu, net, cpuStakedSum, netStakedSum); err != nil {
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
func (s *State) CancelDelegate(from, to common.AccountName, cpuStaked, netStaked uint64) error {
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(from)
	if err != nil {
		return err
	}
	acc.subVotes(cpuStaked + netStaked)
	if from != to {
		accTo, err := s.GetAccountByName(to)
		if err != nil {
			return err
		}
		if err := acc.CancelDelegateOther(accTo, cpuStaked, netStaked, cpuStakedSum, netStakedSum); err != nil {
			return err
		}
		if err := s.CommitAccount(accTo); err != nil {
			return err
		}
	} else {
		if err := acc.CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum); err != nil {
			return err
		}
	}
	value := new(big.Int).Add(new(big.Int).SetUint64(uint64(cpuStaked)), new(big.Int).SetUint64(uint64(netStaked)))
	if err := acc.AddBalance(AbaToken, value); err != nil {
		return err
	}
	amount, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	if err := s.CommitParam(cpuAmount, amount-cpuStaked); err != nil {
		return err
	}
	amount, err = s.GetParam(netAmount)
	if err != nil {
		return err
	}
	if err := s.CommitParam(netAmount, amount-cpuStaked); err != nil {
		return err
	}

	return s.CommitAccount(acc)
}

/**
 *  @brief recover a account's resource by time
 *  @param index - account's index
 *  @param timeStamp - current time
 */
func (s *State) RecoverResources(index common.AccountName, timeStamp int64) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return err
	}
	acc.RecoverResources(cpuStakedSum, netStakedSum, timeStamp)
	return s.CommitAccount(acc)
}

/**
 *  @brief require a account's resource info
 *  @param index - account's index
 *  @param timeStamp - current time
 */
func (s *State) RequireResources(index common.AccountName, timeStamp int64) (float32, float32, error) {
	cpuStakedSum, err := s.GetParam(cpuAmount)
	if err != nil {
		return 0, 0, err
	}
	netStakedSum, err := s.GetParam(netAmount)
	if err != nil {
		return 0, 0, err
	}
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return 0, 0, err
	}
	acc.RecoverResources(cpuStakedSum, netStakedSum, timeStamp)
	log.Debug("cpu:", acc.Cpu.Used, acc.Cpu.Available, acc.Cpu.Limit)
	log.Debug("net:", acc.Net.Used, acc.Net.Available, acc.Net.Limit)
	return acc.Cpu.Available, acc.Net.Available, nil
}

/**
 *  @brief set the cpu and net limits
 *  @param cpu - if true then increase cpu, otherwise, reduce cpu
 *  @param net - if true then increase net, otherwise, reduce net
 */
func (s *State) SetBlockLimits(cpu, net bool) {
	if cpu {
		BlockCpu += BlockCpu * 0.01
		if BlockCpu > VirtualBlockCpuLimit {
			BlockCpu = VirtualBlockCpuLimit
		}
	} else {
		BlockCpu -= BlockCpu * 0.01
		if BlockCpu > BlockCpuLimit {
			BlockCpu = BlockCpuLimit
		}
	}
	if net {
		BlockNet += BlockNet * 0.01
		if BlockNet > VirtualBlockNetLimit {
			BlockNet = VirtualBlockNetLimit
		}
	} else {
		BlockNet -= BlockNet * 0.01
		if BlockNet < BlockNetLimit {
			BlockNet = BlockNetLimit
		}
	}
	log.Debug("SetBlockLimits:", BlockCpu, BlockNet)
}
func (s *State) RegisterProducer(index common.AccountName) error {
	if _, ok := s.Producers[index]; ok {
		return errors.New(log, fmt.Sprintf("the account:%s was already registed", common.IndexToName(index)))
	}
	s.Producers[index] = 0
	return nil
}
func (s *State) PutProducerToVote(index common.AccountName, accounts []common.AccountName) error {
	votingSum, err := s.GetParam(votingAmount)
	if err != nil {
		return err
	}
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if acc.Resource.Votes.Staked == 0 {
		return errors.New(log, fmt.Sprintf("the account:%s has no enough vote", index.String()))
	}
	for _, v := range accounts {
		if _, ok := s.Producers[v]; !ok {
			return errors.New(log, fmt.Sprintf("the account:%s is not register", v.String()))
		}
	}
	acc.UpdateElectedProducers(accounts)
	if err := s.updateProducers(accounts, acc.Resource.Votes.Staked); err != nil {
		return err
	}
	if err := s.CommitParam(votingAmount, votingSum + acc.Resource.Votes.Staked); err != nil {
		return err
	}
	if votingSum + acc.Resource.Votes.Staked > abaTotal * 0.15 {
		log.Warn("Start Process ##################################################################################")
	}
	return s.CommitAccount(acc)
}
func (s *State) updateProducers(accounts []common.AccountName, value uint64) error {
	for _, index := range accounts {
		s.Producers[index] = value
		if err := s.trie.TryUpdate(index.Bytes(), common.Uint64ToBytes(value)); err != nil {
			return err
		}
	}
	return nil
}
func (s *State) RequireVotingInfo() {
	log.Debug(s.Producers)
	votingSum, err := s.GetParam(votingAmount)
	if err != nil {
		log.Error(err)
	}
	log.Debug("votingSum", votingSum, votingSum / abaTotal)
}

/**
 *  @brief set the cpu and net resource to account
 *  @param self - if self, set resource to staked, otherwise, set resource to delegated
 *  @param cpuStaked - stake delegated cpu
 *  @param netStaked - stake delegated net
 *  @param cpuStakedSum - total stake cpu
 *  @param netStakedSum - total stake net
 */
func (a *Account) AddResourceLimits(self bool, cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64) error {
	if self {
		a.Cpu.Staked += cpuStaked
		a.Net.Staked += netStaked
	} else {
		a.Cpu.Delegated += cpuStaked
		a.Net.Delegated += netStaked
	}
	return a.UpdateResource(cpuStakedSum, netStakedSum)
}
func (a *Account) CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64) error {
	a.Cpu.Staked -= cpuStaked
	a.Net.Staked -= netStaked
	return a.UpdateResource(cpuStakedSum, netStakedSum)
}
func (a *Account) CancelDelegateOther(acc *Account, cpuStaked, netStaked, cpuStakedSum, netStakedSum uint64) error {
	done := false
	for i := 0; i < len(a.Delegates); i++ {
		if a.Delegates[i].Index == acc.Index {
			done = true
			if acc.Cpu.Delegated < cpuStaked {
				return errors.New(log, fmt.Sprintf("the account:%s cpu amount is not enough", common.IndexToName(acc.Index)))
			}
			if acc.Net.Delegated < netStaked {
				return errors.New(log, fmt.Sprintf("the account:%s net amount is not enough", common.IndexToName(acc.Index)))
			}
			acc.CancelDelegateSelf(cpuStaked, netStaked, cpuStakedSum, netStakedSum)

			a.Delegates[i].CpuStaked -= cpuStaked
			a.Delegates[i].NetStaked -= netStaked
			if a.Delegates[i].CpuStaked == 0 && a.Delegates[i].NetStaked == 0 {
				a.Delegates = append(a.Delegates[:i], a.Delegates[i+1:]...)
			}
		}
	}
	if done == false {
		return errors.New(log, fmt.Sprintf("account:%s is not delegated for %s", common.IndexToName(a.Index), common.IndexToName(acc.Index)))
	}
	return nil
}
func (a *Account) SubResourceLimits(cpu, net float32, cpuStakedSum, netStakedSum uint64) error {
	if a.Cpu.Available < cpu {
		return errors.New(log, fmt.Sprintf("the account:%s cpu amount is not enough", common.IndexToName(a.Index)))
	}
	if a.Net.Available < net {
		return errors.New(log, fmt.Sprintf("the account:%s net amount is not enough", common.IndexToName(a.Index)))
	}
	a.Cpu.Used += cpu
	a.Net.Used += net
	return a.UpdateResource(cpuStakedSum, netStakedSum)
}
func (a *Account) SetDelegateInfo(index common.AccountName, cpuStaked, netStaked uint64) error {
	d := Delegate{Index: index, CpuStaked: cpuStaked, NetStaked: netStaked}
	a.Delegates = append(a.Delegates, d)
	return nil
}
func (a *Account) UpdateResource(cpuStakedSum, netStakedSum uint64) error {
	if cpuStakedSum == 0 || netStakedSum == 0 {
		return errors.New(log, "cpuStakedSum and netStakedSum can't be zero")
	}
	a.Cpu.Limit = float32(a.Cpu.Staked+a.Cpu.Delegated) / float32(cpuStakedSum) * BlockCpu
	a.Cpu.Available = a.Cpu.Limit - a.Cpu.Used
	a.Net.Limit = float32(a.Cpu.Staked+a.Net.Delegated) / float32(netStakedSum) * BlockNet
	a.Net.Available = a.Net.Limit - a.Net.Used
	return nil
}
func (a *Account) RecoverResources(cpuStakedSum, netStakedSum uint64, timeStamp int64) error {
	t := timeStamp / (1000 * 1000)
	interval := 100.0 * float32(t-a.TimeStamp) / (24.0 * 60.0 * 60.0 * 1000)
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
	a.UpdateResource(cpuStakedSum, netStakedSum)
	a.TimeStamp = t
	return nil
}
func (a *Account) addVotes(token uint64) {
	a.Resource.Votes.Staked += token
}
func (a *Account) subVotes(token uint64) {
	a.Resource.Votes.Staked -= token
}

func (a *Account) UpdateElectedProducers(accounts []common.AccountName) error {
	for k := range a.Resource.Votes.Producers {
		delete(a.Votes.Producers, k)
	}
	for _, v := range accounts {
		a.Resource.Votes.Producers[v] = a.Resource.Votes.Staked
	}
	return nil
}
