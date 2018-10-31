package settlement

import (
	"errors"
	"bytes"
	"context"
	"io/ioutil"
	"math/big"
	"time"
	dproof "github.com/ecoball/go-ecoball/dsn/proof"
	"github.com/ecoball/go-ecoball/dsn/host"
	"github.com/ecoball/go-ecoball/dsn/common/ecoding"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/core/ledgerimpl/ledger"
	"github.com/ecoball/go-ecoball/core/state"
	ecommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/dsn/ipfs/api"
	"github.com/ecoball/go-ecoball/common/elog"
	dsnComm "github.com/ecoball/go-ecoball/dsn/common"
	"github.com/ecoball/go-ecoball/core/types"
	innerCommon "github.com/ecoball/go-ecoball/common"
	"github.com/ecoball/go-ecoball/common/event"
	"github.com/ecoball/go-ecoball/common/config"
)

var (
	errProofInvalid = errors.New("Storage proof is invalid")
	log = elog.NewLogger("dsn-s", elog.DebugLog)
)

type Settler struct {
	ledger    ledger.Ledger
	chainId   string
	ctx       context.Context
}

func NewStorageSettler(ctx context.Context, l ledger.Ledger, chainId string) (*Settler, error) {
	return &Settler{
		ledger: l,
		chainId: chainId,
		ctx:ctx,
	}, nil
}

func (s *Settler) Start() error {
	//s.rxLoop()
	return nil
}

func (s *Settler) payToHost(spf host.StorageProof, st state.InterfaceState) error {
	reward := CalcHostReward(spf, st)
	timeNow := time.Now().UnixNano()
	tran, err := types.NewTransfer(ecommon.NameToIndex(dsnComm.RootAccount),
		innerCommon.NameToIndex(spf.AccountName), ecommon.HexToHash(s.chainId), "owner", reward, 0, timeNow)
	if err != nil {
		return err
	}
	tran.SetSignature(&config.Root)
	err = event.Send(event.ActorNil, event.ActorTxPool, tran)
	if err != nil {
		return err
	}
	return nil
}

// decodeAnnouncement decodes announcement bytes into a host announcement
func (s *Settler) decodeAnnouncement(fullAnnouncement []byte) (host.HostAncContract, error) {
	var announcement host.HostAncContract
	dec := encoding.NewDecoder(bytes.NewReader(fullAnnouncement))
	err := dec.Decode(&announcement)
	if err != nil {
		return announcement, err
	}
	/*var sig [dsnComm.SigSize]byte
	err = dec.Decode(&sig)
	if err != nil {
		return announcement, err
	}
	anHash := sha256.Sum256(encoding.Marshal(announcement))
	ok, err := secp256k1.Verify(anHash[:], sig[:], announcement.PublicKey)
	if !ok {
		log.Error("sig check failed")
	}
	if err != nil {
		return announcement, err
	}*/
	return announcement, nil
}

func (s *Settler) decodeProof(proof []byte) (host.StorageProof, error) {
	var sp host.StorageProof
	dec := encoding.NewDecoder(bytes.NewReader(proof))
	err := dec.Decode(&sp)
	if err != nil {
		return sp, err
	}
	/*var sig [dsnComm.SigSize]byte
	err = dec.Decode(&sig)
	if err != nil {
		return sp, err
	}
	proofHash := sha256.Sum256(encoding.Marshal(sp))
	ok, err := secp256k1.Verify(proofHash[:], sig[:], sp.PublicKey)
	if !ok {
		log.Error("sig check failed")
	}
	if err != nil {
		return sp, err
	}*/
	return sp, nil
}

func (s *Settler) decodeFileContract(data []byte) (renter.FileContract, error) {
	var fc renter.FileContract
	dec := encoding.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&fc)
	if err != nil {
		return fc, err
	}
	/*var sig [dsnComm.SigSize]byte
	err = dec.Decode(&sig)
	if err != nil {
		return fc, err
	}
	fcHash := sha256.Sum256(encoding.Marshal(fc))
	ok, err := secp256k1.Verify(fcHash[:], sig[:], fc.PublicKey)
	if !ok {
		log.Error("sig check failed")
	}
	if err != nil {
		return fc, err
	}*/
	return fc, nil
}

func (s *Settler)verifyStorageProof(proof host.StorageProof, st state.InterfaceState) (bool, error) {
	block, err := api.IpfsBlockGet(s.ctx, proof.Cid)
	if err != nil {
		return false, err
	}
	blockData, err := ioutil.ReadAll(block)
	if err != nil {
		return false, err
	}
	rootHash := dproof.MerkleRoot(blockData)
	numberSegment := len(blockData) / dproof.SegmentSize
	ret := dproof.VerifySegment(proof.Segment[:], proof.HashSet, uint64(numberSegment), proof.SegmentIndex, rootHash)
	return ret, err
}

func (s *Settler) storeAccountState(data interface{}, st state.InterfaceState) error {
	var err error
	switch data.(type) {
	case *host.HostAncContract:
		value := data.(*host.HostAncContract)
		err = updateStateHostAnn(value, st)
	case *host.StorageProof:
		value := data.(*host.StorageProof)
		err = updateStateProof(value, st)
	case *renter.FileContract:
		value := data.(*renter.FileContract)
		err = updateRenterFiles(value, st)
	default:
		err = errors.New("unknowed data type")
	}
	return err
}

func (s *Settler)HandleHostAnce(data []byte, st state.InterfaceState) error {
	c, err := s.decodeAnnouncement(data)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return s.storeAccountState(c,st)
}

func (s *Settler)HandleStorageProof(data []byte, st state.InterfaceState) error {
	proof, err := s.decodeProof(data)
	if err != nil {
		return err
	}
	valid, err := s.verifyStorageProof(proof, st)
	if err != nil {
		return err
	}
	if !valid {
		return errProofInvalid
	}
	s.payToHost(proof, st)
	s.storeAccountState(proof, st)
	return nil
}

func (s *Settler) HandleFileContract(data []byte, st state.InterfaceState) error {
	fc, err := s.decodeFileContract(data)
	if err != nil {
		return err
	}
	s.storeAccountState(fc, st)
	//PinBlock(fc.Cid)
	return nil
}

func updateStateHostAnn(an *host.HostAncContract, st state.InterfaceState) error {
	var exsitHost bool
	var hc HostAnceSource
	sKey := []byte(KeyStorageAn)
	hbuff, err := st.StoreGet(ecommon.NameToIndex(an.AccountName), sKey)
	if err == nil {
		exsitHost = true
		encoding.Unmarshal(hbuff, &hc)
	}

	var du DiskResource
	dkey := []byte(KeyStorageTotal)
	dbuff, _ := st.StoreGet(ecommon.NameToIndex(dsnComm.RootAccount), dkey)
	encoding.Unmarshal(dbuff, &du)
	if exsitHost {
		du.TotalCapacity = du.TotalCapacity - hc.TotalStorage + an.TotalStorage
	} else {
		du.TotalCapacity = du.TotalCapacity + an.TotalStorage
		du.Hosts = append(du.Hosts, an.AccountName)
	}
	newDbuff := encoding.Marshal(du)
	st.StoreSet(ecommon.NameToIndex(dsnComm.RootAccount), dkey, newDbuff)

	vb := encoding.Marshal(HostAnceSource{
		TotalStorage: an.TotalStorage,
		StartAt: an.StartAt,
	})
	return st.StoreSet(ecommon.NameToIndex(an.AccountName), sKey, vb)
}

func updateStateProof(sp *host.StorageProof, st state.InterfaceState) error {
	var exsit bool
	var oldSp ProofInfo
	spKey := []byte(KeyStorageProof)
	oldBuf, err := st.StoreGet(ecommon.NameToIndex(sp.AccountName), spKey)
	if err == nil {
		exsit = true
		encoding.Unmarshal(oldBuf, &oldSp)
	}

	var du DiskResource
	dkey := []byte(KeyStorageTotal)
	dbuff, _ := st.StoreGet(ecommon.NameToIndex(dsnComm.RootAccount), dkey)
	encoding.Unmarshal(dbuff, &du)
	if exsit {
		du.UsedCapacity = du.UsedCapacity - oldSp.RepoSize + sp.RepoSize
	} else {
		du.UsedCapacity = du.UsedCapacity + sp.RepoSize
	}
	newDuBuff := encoding.Marshal(du)
	st.StoreSet(ecommon.NameToIndex(dsnComm.RootAccount), dkey, newDuBuff)

	newSp := oldSp
	newSp.RepoSize = sp.RepoSize
	newSp.Snapshot = append(newSp.Snapshot, sp.AtHeight)
	newBuf := encoding.Marshal(newSp)
	return st.StoreSet(ecommon.NameToIndex(sp.AccountName), spKey, newBuf)
}

func updateRenterFiles(fc *renter.FileContract, st state.InterfaceState) error {
	var dr DiskResource
	totalKey := []byte(KeyStorageTotal)
	drBuff, _ := st.StoreGet(ecommon.NameToIndex(dsnComm.RootAccount), totalKey)
	encoding.Unmarshal(drBuff, &dr)
	dr.TotalFileSize = dr.TotalFileSize + fc.FileSize
	dr.TotalFileCount++
	newDrBuf := encoding.Marshal(dr)
	st.StoreSet(ecommon.NameToIndex(dsnComm.RootAccount), totalKey, newDrBuf)

	var fs Files
	fsKey := []byte(KeyStorageFile)
	fsBuff, _ := st.StoreGet(ecommon.NameToIndex(fc.AccountName), fsKey)
	encoding.Unmarshal(fsBuff, &fs)
	newFile := fileInfo{
		FileSize:fc.FileSize,
		Redundancy: fc.Redundancy,
	}
	fs.AllFiles = append(fs.AllFiles, newFile)
	newFsBuf := encoding.Marshal(fs)
	st.StoreSet(ecommon.NameToIndex(fc.AccountName), fsKey, newFsBuf)
	return nil
}

func CalcHostReward(spf host.StorageProof, st state.InterfaceState) *big.Int {
	var dr DiskResource
	totalKey := []byte(KeyStorageTotal)
	drBuff, _ := st.StoreGet(ecommon.NameToIndex(dsnComm.RootAccount), totalKey)
	encoding.Unmarshal(drBuff, &dr)

	var hn HostAnceSource
	anKey := []byte(KeyStorageAn)
	hnBuff, _ := st.StoreGet(ecommon.NameToIndex(spf.AccountName), anKey)
	encoding.Unmarshal(hnBuff, &hn)

	var pi ProofInfo
	piKey := []byte(KeyStorageProof)
	piBuff, _ := st.StoreGet(ecommon.NameToIndex(spf.AccountName), piKey)
	encoding.Unmarshal(piBuff, &pi)

	var ontime int = 1
	shotCnt := len(pi.Snapshot)
	for i:= shotCnt -1; i > 0; i-- {
		interval := pi.Snapshot[i] - pi.Snapshot[i - 1]
		if interval  > BlkHeightPerDay - BlkHeightPerHour && interval < BlkHeightPerDay + BlkHeightPerHour {
			ontime++
		} else {
			break
		}
	}

	reward := (float64(hn.TotalStorage) / float64(dr.UsedCapacity)) * RatioTotal + float64(spf.RepoSize) * RatioUsed + float64(ontime) * RationOntime

	var reBit  big.Int
	return reBit.SetInt64(int64(reward * 100000))
}


