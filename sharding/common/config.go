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
	DefaultSyncBlockTimer = 1
	DefaultRetransTimer   = 1
	DefaultFullVoteTimer  = 3

	DefaultProductCmBlockTimer = 120  //second
	DefaultCmBlockWindow       = 1000 //Millisecond

	DefaultProductFinalBlockTimer = 180  //second
	DefaultFinalBlockWindow       = 1000 //Millisecond

	DefaultProductViewChangeBlockTimer = 180  //second
	DefaultViewchangeBlockWindow       = 1000 //Millisecond

	DefaultWaitMinorBlockTimer  = 180  //second
	DefaultWaitMinorBlockWindow = 10   //second
	DefaultMinorBlockWindow     = 1000 //Millisecond

	DefaultBlockWindow = 0 //Millisecond
)
