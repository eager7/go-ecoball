package message

type GetTxs struct{}

type ABABFTStart struct{}
type SoloStop struct {}
type GetCurrentHeader struct{}

type GetTransaction struct {
	Key []byte
}
