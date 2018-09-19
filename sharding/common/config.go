package common

//DO NOT EDIT
const (
	NodeCommittee = iota + 1
	NodeShard
	NodeLookup
	NodeCandidate
	NodeNil
)

//DO NOT EDIT
const (
	DefaultCommitteMaxMember     = 100
	DefaultShardMaxMember        = 100
	DefaultEpochFinalBlockNumber = 100

	DefaultThresholdOfMinorBlock = 80  /*Percent*/
	DefaultThresholdOfConsensus  = 400 /*thousandth*/

	/*timer of fsm state .second*/
	DefaultSyncBlockTimer = 10 //180
	DefaultRetransTimer   = 1  //180

	DefaultBlockWindow = 2

	DefaultProductCmBlockTimer = 20
	DefaultCmBlockWindow       = 10

	DefaultWaitMinorBlockTimer    = 20
	DefaultProductFinalBlockTimer = 20
	DefaultFinalBlockWindow       = 10

	DefaultProductViewChangeBlockTimer = 40
)
