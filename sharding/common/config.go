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

	DefaultProductCmBlockTimer = 120  //second
	DefaultCmBlockWindow       = 1000 //Millisecond, time to product block

	DefaultProductFinalBlockTimer = 180  //second
	DefaultFinalBlockWindow       = 1000 //Millisecond, time to product block

	DefaultProductViewChangeBlockTimer = 180  //second
	DefaultViewchangeBlockWindow       = 1000 //Millisecond, time to product block

	DefaultWaitMinorBlockTimer  = 180  //second, time to collect shard blocks
	DefaultWaitMinorBlockWindow = 12   //second, time wait for all shard blocks,  must be thrice of FullVoteTimer
	DefaultMinorBlockWindow     = 1000 //Millisecond, time to product block

	DefaultBlockWindow = 0 //Millisecond
)
