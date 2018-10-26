package state

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/crypto/secp256k1"
	errIn "errors"
)

var Owner = "owner"
var Active = "active"

type AccFactor struct {
	Actor      common.AccountName `json:"actor"`
	Weight     uint32             `json:"weight"`
	Permission string             `json:"permission"`
}

type KeyFactor struct {
	Actor  common.Address `json:"actor"`
	Weight uint32         `json:"weight"`
}

type Permission struct {
	PermName  string               `json:"perm_name"`
	Parent    string               `json:"parent"`
	Threshold uint32               `json:"threshold"`
	Keys      map[string]KeyFactor `json:"keys, omitempty"`     //map[key's hex string]KeyFactor
	Accounts  map[string]AccFactor `json:"accounts, omitempty"` //map[actor's string]AccFactor
}

/**
 *  @brief create a new permission object
 *  @param name - the permission's name
 *  @param parent - the parent name of this permission, if the permission's name is 'owner', then the parent is null
 *  @param threshold - the threshold of this permission, when the weight greater than or equal to threshold, permission will only take effect
 *  @param addr - the public keys list
 *  @param acc - the accounts list
 */
func NewPermission(name, parent string, threshold uint32, addr []KeyFactor, acc []AccFactor) Permission {
	Keys := make(map[string]KeyFactor, 1)
	for _, a := range addr {
		Keys[a.Actor.HexString()] = a
	}
	Accounts := make(map[string]AccFactor, 1)
	for _, a := range acc {
		Accounts[a.Actor.String()] = a
	}
	return Permission{
		PermName:  name,
		Parent:    parent,
		Threshold: threshold,
		Keys:      Keys,
		Accounts:  Accounts,
	}
}

/**
 *  @brief check that the signatures meets the permission requirement
 *  @param state - the mpt trie, used to search account
 *  @param signatures - the transaction's signatures list
 */
func (p *Permission) checkPermission(state *State, signatures []common.Signature) error {
	Keys := make(map[common.Address][]byte, 1)
	Accounts := make(map[string][]byte, 1)
	for _, s := range signatures {
		addr := common.AddressFromPubKey(s.PubKey)
		acc, err := state.GetAccountByAddr(addr)
		if err == nil {
			Accounts[acc.Index.String()] = s.SigData
		} else {
			log.Warn("permission", p.PermName, "error:", err) //allow mixed with invalid account, just have enough weight
		}
		Keys[addr] = s.SigData
	}
	var weightKey uint32
	for addr := range Keys {
		if key, ok := p.Keys[addr.HexString()]; ok {
			weightKey += key.Weight
		}
		if weightKey >= p.Threshold {
			return nil
		}
	}
	var weightAcc uint32
	for acc := range Accounts {
		if a, ok := p.Accounts[acc]; ok {
			weightAcc += a.Weight
			if next, err := state.GetAccountByName(a.Actor); err != nil {
				return err
			} else {
				perm := next.Permissions[a.Permission]
				if err := perm.checkPermission(state, signatures); err != nil {
					return err
				}
			}
		}
		if weightAcc >= p.Threshold {
			return nil
		}
	}

	return errors.New(log, fmt.Sprintf("weight is not enough, keys weight:%d, accounts weight:%d", weightKey, weightAcc))
}

/**
 *  @brief check if guest has host permission, this method will not modified mpt trie
 *  @param state - the mpt trie, used to search account
 *  @param guest - the guest account
 *  @param permission - the permission names
 */
func (p *Permission) checkAccountPermission(state *State, guest string, permission string) error {
	var weightAcc uint32
	if a, ok := p.Accounts[guest]; ok {
		weightAcc += a.Weight
		if _, err := state.GetAccountByName(a.Actor); err != nil {
			return err
		}
	}
	if weightAcc >= p.Threshold {
		return nil
	}

	return errIn.New(fmt.Sprintf("%s@%s  weight is not enough, accounts weight:%d", guest, permission, weightAcc))
}

/**
 *  @brief add a permission object into account, then update to mpt trie
 *  @param perm - the permission object
 */
func (s *State) AddPermission(index common.AccountName, perm Permission) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	acc.AddPermission(perm)
	return s.commitAccount(acc)
}

/**
 *  @brief check the permission's validity, this method will not modified mpt trie
 *  @param index - the account index
 *  @param state - the world state tree
 *  @param name - the permission names
 *  @param signatures - the signatures list
 */
func (s *State) CheckPermission(index common.AccountName, name string, hash common.Hash, signatures []common.Signature) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	var sig []common.Signature
	for _, v := range signatures {
		result, err := secp256k1.Verify(hash.Bytes(), v.SigData, v.PubKey)
		if err == nil && result == true {
			sig = append(sig, v)
		} else {
			log.Warn("verify signature failed:", err, result)
		}
	}
	return acc.checkPermission(s, name, sig)
}

/**
 *  @brief check if guest has host permission, this method will not modified mpt trie
 *  @param host - the host account
 *  @param guest - the guest account
 *  @param permission - the permission names
 */
func (s *State) CheckAccountPermission(host common.AccountName, guest common.AccountName, permission string) error {
	acc, err := s.GetAccountByName(host)
	if err != nil {
		return err
	}
	acc.mutex.Lock()
	defer acc.mutex.Unlock()
	return acc.checkAccountPermission(s, guest.String(), permission)
}

/**
 *  @brief search the permission by name, return json array string
 *  @param index - the account index
 *  @param name - the permission names
 */
func (s *State) FindPermission(index common.AccountName, name string) (string, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return "", err
	}
	acc.mutex.RLock()
	defer acc.mutex.RUnlock()
	if str, err := acc.findPermission(name); err != nil {
		return "", err
	} else {
		return "[" + str + "]", nil
	}
}

/**
 *  @brief set the permission into account, if the permission existed, will be to overwrite
 *  @param name - the permission name
 */
func (a *Account) AddPermission(perm Permission) {
	a.Permissions[perm.PermName] = perm
}

/**
 *  @brief check that the signatures meets the permission requirement
 *  @param state - the mpt trie, used to search account
 *  @param name - the permission name
 *  @param signatures - the transaction's signatures list
 */
func (a *Account) checkPermission(state *State, name string, signatures []common.Signature) error {
	if perm, ok := a.Permissions[name]; !ok {
		return errors.New(log, fmt.Sprintf("can't find this permission in account:%s", name))
	} else {
		if "" != perm.Parent {
			if err := a.checkPermission(state, perm.Parent, signatures); err == nil {
				return nil
			}
		}
		if err := perm.checkPermission(state, signatures); err != nil {
			log.Error(fmt.Sprintf("account:%s", a.JsonString(false)))
			return err
		}
	}
	return nil
}

/**
 *  @brief check if guest has host permission, this method will not modified mpt trie
 *  @param state - the mpt trie, used to search account
 *  @param guest - the guest account
 *  @param permission - the permission names
 */
func (a *Account) checkAccountPermission(state *State, guest string, permission string) error {
	if perm, ok := a.Permissions[permission]; !ok {
		return errors.New(log, fmt.Sprintf("account %s has not %s permission of account:%s", guest, permission, a.Index.String()))
	} else {
		if "" != perm.Parent {
			if err := a.checkAccountPermission(state, guest, perm.Parent); err == nil {
				return nil
			}
		}
		if err := perm.checkAccountPermission(state, guest, permission); err != nil {
			log.Error(fmt.Sprintf("account:%s", a.JsonString(false)))
			return err
		}
	}
	return nil
}

/**
 *  @brief get the permission information by name, return json string
 *  @param name - the permission name
 */
func (a *Account) findPermission(name string) (str string, err error) {
	perm, ok := a.Permissions[name]
	if !ok {
		return "", errors.New(log, fmt.Sprintf("can't find this permission:%s", name))
	}
	b, err := json.Marshal(perm)
	if err != nil {
		return "", err
	}
	str += string(b)
	if "" != perm.Parent {
		if s, err := a.findPermission(perm.Parent); err == nil {
			str += "," + s
		}
	}
	return string(str), nil
}
