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
	DefaultThresholdOfConsensus  = 667 /*thousandth*/

	/*timer of fsm state .second*/
	DefaultSyncBlockTimer = 10 //180
	DefaultRetransTimer   = 1  //180

	DefaultProductCmBlockTimer = 60
	DefaultCmBlockWindow       = 10

	DefaultProductFinalBlockTimer = 60
	DefaultFinalBlockWindow       = 10

	DefaultProductViewChangeBlockTimer = 180

	DefaultWaitMinorBlockTimer = 180
)
