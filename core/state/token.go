package state

import (
	"encoding/json"
	"fmt"
	"github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/errors"
	"github.com/ecoball/go-ecoball/core/pb"
	"math/big"
)

const AbaTotal = 2100000000

type TokenInfo struct {
	Symbol    string             `json:"symbol"`
	MaxSupply *big.Int           `json:"max_supply"`
	Supply    *big.Int           `json:"supply"`
	Creator   common.AccountName `json:"issuer"`
	Issuer    common.AccountName `json:"issuer"`
}

type Token struct {
	Name    string   `json:"index"`
	Balance *big.Int `json:"balance, omitempty"`
}

func NewToken(symbol string, maxSupply, supply *big.Int, creator, issuer common.AccountName) (*TokenInfo, error) {
	stat := &TokenInfo{
		Symbol:    symbol,
		MaxSupply: maxSupply,
		Supply:    supply,
		Creator:   creator,
		Issuer:    issuer,
	}

	return stat, nil
}

func (info *TokenInfo) Serialize() ([]byte, error) {
	maxSupply, err := info.MaxSupply.GobEncode()
	supply, err := info.Supply.GobEncode()
	p := &pb.TokenInfo{
		Symbol:    info.Symbol,
		MaxSupply: maxSupply,
		Supply:    supply,
		Creator:   uint64(info.Creator),
		Issuer:    uint64(info.Issuer),
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (info *TokenInfo) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data's length is zero")
	}
	var status pb.TokenInfo
	if err := status.Unmarshal(data); err != nil {
		return err
	}

	maxSupply := new(big.Int)
	if err := maxSupply.GobDecode(status.MaxSupply); err != nil {
		return errors.New(fmt.Sprintf("GobDecode err:%s", err.Error()))
	}

	supply := new(big.Int)
	if err := supply.GobDecode(status.Supply); err != nil {
		return errors.New(fmt.Sprintf("GobDecode err:%s", err.Error()))
	}

	info.Symbol = status.Symbol
	info.MaxSupply = maxSupply
	info.Supply = supply
	info.Creator = common.AccountName(status.Creator)
	info.Issuer = common.AccountName(status.Issuer)

	return nil
}

func (info *TokenInfo) JsonString(format bool) string {
	if format {
		data, err := json.MarshalIndent(info, "", "    ")
		if err != nil {
			fmt.Println(err)
		}
		return string(data)
	} else {
		data, err := json.Marshal(info)
		if err != nil {
			fmt.Println(err)
		}
		return string(data)
	}
}

func (s *State) AccountGetBalance(index common.AccountName, token string) (*big.Int, error) {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return nil, err
	}
	acc.lock.RLock()
	defer acc.lock.RUnlock()
	return acc.Balance(token)
}
func (s *State) AccountSubBalance(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	balance, err := acc.Balance(token)
	if err != nil {
		return err
	}
	if balance.Cmp(value) == -1 {
		return errors.New("no enough balance")
	}
	if err := acc.SubBalance(token, value); err != nil {
		return err
	}
	if err := s.CommitAccount(acc); err != nil {
		return err
	}
	return nil
}
func (s *State) AccountAddBalance(index common.AccountName, token string, value *big.Int) error {
	acc, err := s.GetAccountByName(index)
	if err != nil {
		return err
	}
	if !s.TokenExisted(token) {
		return errors.New(fmt.Sprintf("%s token is not existed", token))
	}
	acc.lock.Lock()
	defer acc.lock.Unlock()
	if err := acc.AddBalance(token, value); err != nil {
		return err
	}
	if err := s.CommitAccount(acc); err != nil {
		return err
	}

	return nil
}

func (s *State) TokenExisted(name string) bool {
	if token := s.Tokens.Get(name); token != nil {
		return true
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	data, err := s.trie.TryGet([]byte(name))
	if err != nil {
		log.Error(err)
		return false
	}

	if data == nil {
		return false
	}

	token := &TokenInfo{}
	if err = token.Deserialize(data); err != nil {
		return false
	}

	return token.Symbol == name
}

func (s *State) GetTokenInfo(symbol string) (*TokenInfo, error) {
	if err := common.TokenNameCheck(symbol); err != nil {
		return nil, errors.New(err.Error())
	}

	if token := s.Tokens.Get(symbol); token != nil {
		return token, nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	data, err := s.trie.TryGet([]byte(symbol))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if data == nil {
		return nil, errors.New(fmt.Sprintf("no this token named:%s", symbol))
	}

	token := &TokenInfo{}
	if err := token.Deserialize(data); err != nil {
		return nil, err
	}
	return token, nil
}

/**
 *  @brief update the account's information into trie
 *  @param acc - account object
 */
func (s *State) CommitToken(token *TokenInfo) error {
	if token == nil {
		return errors.New("param acc is nil")
	}
	d, err := token.Serialize()
	if err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.trie.TryUpdate([]byte(token.Symbol), d); err != nil {
		return err
	}
	s.Tokens.Add(token.Symbol, *token)
	return nil
}

func (s *State) CreateToken(symbol string, maxSupply *big.Int, creator, issuer common.AccountName) (*TokenInfo, error) {
	if err := common.TokenNameCheck(symbol); err != nil {
		return nil, err
	}

	if s.TokenExisted(symbol) {
		return nil, errors.New(fmt.Sprintf("%s token had created", symbol))
	}

	token, err := NewToken(symbol, maxSupply, big.NewInt(0), creator, issuer)
	if err != nil {
		return nil, err
	}

	if err := s.CommitToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// for token contract api
func (s *State) SetTokenInfo(symbol string, maxSupply, supply *big.Int, creator, issuer common.AccountName) (*TokenInfo, error) {
	if err := common.TokenNameCheck(symbol); err != nil {
		return nil, err
	}

	token, err := NewToken(symbol, maxSupply, supply, creator, issuer)
	if err != nil {
		return nil, err
	}

	if err := s.CommitToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *State) IssueToken(to common.AccountName, amount *big.Int, symbol string) error {
	token, err := s.GetTokenInfo(symbol)
	if err != nil {
		return err
	}

	balance := new(big.Int).Sub(token.MaxSupply, token.Supply)

	if balance.Cmp(amount) == -1 {
		return errors.New("no enough balance")
	}

	if err = s.AccountAddBalance(to, symbol, amount); err != nil {
		return err
	}

	token.Supply = new(big.Int).Add(token.Supply, amount)

	if err := s.CommitToken(token); err != nil {
		return err
	}

	return nil
}

/**
 *  @brief create a new token in account
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (a *Account) AddToken(name string) error {
	log.Debug("add token:", name)
	ac := Token{Name: name, Balance: new(big.Int).SetUint64(0)}
	a.Tokens[name] = ac
	return nil
}

/**
 *  @brief check the token for existence, return true if existed
 *  @param index - the unique id of token name created by common.NameToIndex()
 */
func (a *Account) TokenExisted(token string) bool {
	_, ok := a.Tokens[token]
	if ok {
		return true
	}
	return false
}

/**
 *  @brief add balance into account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @param amount - value of token
 */
func (a *Account) AddBalance(name string, amount *big.Int) error {
	//log.Debug("add token", name, "balance:", amount, a.Index)
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	ac, ok := a.Tokens[name]
	if !ok {
		if err := a.AddToken(name); err != nil {
			return err
		}
		ac, _ = a.Tokens[name]
	}
	ac.SetBalance(new(big.Int).Add(ac.GetBalance(), amount))
	a.Tokens[name] = ac
	return nil
}

/**
 *  @brief sub balance into account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @param amount - value of token
 */
func (a *Account) SubBalance(token string, amount *big.Int) error {
	if amount.Sign() == 0 {
		return errors.New("amount is zero")
	}
	t, ok := a.Tokens[token]
	if !ok {
		return errors.New(fmt.Sprintf("account:%s no this token:%s", a.Index.String(), token))
	}
	balance := t.GetBalance()
	value := new(big.Int).Sub(balance, amount)
	if value.Sign() < 0 {
		return errors.New("the balance is not enough")
	}
	t.SetBalance(value)
	a.Tokens[token] = t
	return nil
}

/**
 *  @brief get the balance of account
 *  @param index - the unique id of token name created by common.NameToIndex()
 *  @return big.int - value of token
 */
func (a *Account) Balance(token string) (*big.Int, error) {
	t, ok := a.Tokens[token]
	if !ok {
		return nil, errors.New(fmt.Sprintf("the:%s balance is zero, in account:%s", token, a.Index.String()))
	}
	return t.GetBalance(), nil
}

/**
 *  @brief set balance of account
 *  @param amount - value of token
 */
func (t *Token) SetBalance(amount *big.Int) {
	//TODO:将变动记录存到日志文件
	t.setBalance(amount)
}
func (t *Token) setBalance(amount *big.Int) {
	t.Balance = amount
}
func (t *Token) GetBalance() *big.Int {
	return t.Balance
}
