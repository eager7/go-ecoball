package erasure

import (
	"io"
)


// An ErasureCoder is an error-correcting encoder and decoder.
type ErasureCoder interface {
	// NumPieces is the number of pieces returned by Encode.
	NumPieces() int

	// MinPieces is the minimum number of pieces that must be present to
	// recover the original data.
	MinPieces() int

	// Encode splits data into equal-length pieces, with some pieces
	// containing parity data.
	Encode(data []byte) ([][]byte, error)

	// EncodeShards encodes the input data like Encode but accepts an already
	// sharded input.
	EncodeShards(data [][]byte) ([][]byte, error)

	// Recover recovers the original data from pieces and writes it to w.
	// pieces should be identical to the slice returned by Encode (length and
	// order must be preserved), but with missing elements set to nil. n is
	// the number of bytes to be written to w; this is necessary because
	// pieces may have been padded with zeros during encoding.
	Recover(pieces [][]byte, n uint64, w io.Writer) error


	Parity() [][]byte
}
