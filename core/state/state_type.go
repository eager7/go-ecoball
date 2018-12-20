package state

type TypeState uint8

const (
	FinalType TypeState = 1
	TempType  TypeState = 2
	CopyType  TypeState = 3
)

func (t TypeState) String() string {
	switch t {
	case FinalType:
		return "FinalType"
	case TempType:
		return "TempType"
	case CopyType:
		return "CopyType"
	}
	return "unknown type"
}
