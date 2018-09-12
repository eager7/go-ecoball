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

	DefaultThresholdOfMinorBlock = 80 /*Percent*/

	/*timer of fsm state .second*/
	DefaultSyncBlockTimer              = 10 //180
	DefaultProductCMBlockTimer         = 10 //60
	DefaultWaitMinorBlockTimer         = 10 //120
	DefaultProductFinalBlockTimer      = 10 //60
	DefaultProductViewChangeBlockTimer = 10 //60
)
