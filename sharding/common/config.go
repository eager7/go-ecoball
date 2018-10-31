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
	DefaultSyncBlockTimer = 10
	DefaultRetransTimer   = 1
	DefaultFullVoteTimer  = 3

	DefaultProductCmBlockTimer = 60  //second
	DefaultCmBlockWindow       = 400 //Millisecond

	DefaultProductFinalBlockTimer = 60  //second
	DefaultFinalBlockWindow       = 400 //Millisecond

	DefaultProductViewChangeBlockTimer = 180 //second
	DefaultViewchangeBlockWindow       = 800 //Millisecond

	DefaultWaitMinorBlockTimer  = 180 //second
	DefaultWaitMinorBlockWindow = 10  //second
	DefaultMinorBlockWindow     = 400 //Millisecond

	DefaultBlockWindow = 0 //Millisecond
)
