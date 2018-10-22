package TBLS

import "github.com/ecoball/go-ecoball/common/elog"
var (
	log = elog.NewLogger("tbls.dkg", elog.DebugLog)
)

var privatePoly PriPoly
func StartDKG(index int, threshold int) {
	privatePoly = *SetPriShare(index, threshold)
	go dkgRoutine()
	//c.pvcRoutine()
}
func dkgRoutine() {
	log.Debug("start committee routine")
	/*
	c.stateTimer = time.NewTimer(sc.DefaultSyncBlockTimer * time.Second)
	c.retransTimer = time.NewTimer(sc.DefaultRetransTimer * time.Millisecond)

	for {
		select {
		case msg := <-c.actorc:
			c.processActorMsg(msg)
		case packet := <-c.ppc:
			c.processPacket(packet)
		case <-c.stateTimer.C:
			c.processStateTimeout()
		case <-c.retransTimer.C:
			c.processRetransTimeout()
		}
	}
	*/

}
