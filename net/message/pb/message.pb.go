// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: message.proto

/*
	Package pb is a generated protocol buffer package.

	It is generated from these files:
		message.proto

	It has these top-level messages:
		Message
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// ecoball related message
type MsgType int32

const (
	MsgType_APP_MSG_TRN              MsgType = 0
	MsgType_APP_MSG_BLK              MsgType = 1
	MsgType_APP_MSG_SIGNPRE          MsgType = 2
	MsgType_APP_MSG_BLKF             MsgType = 3
	MsgType_APP_MSG_SIGNBLKF         MsgType = 4
	MsgType_APP_MSG_BLKS             MsgType = 5
	MsgType_APP_MSG_REQSYN           MsgType = 6
	MsgType_APP_MSG_REQSYNSOLO       MsgType = 7
	MsgType_APP_MSG_BLKSYN           MsgType = 8
	MsgType_APP_MSG_TIMEOUT          MsgType = 9
	MsgType_APP_MSG_SHARDING_PACKET  MsgType = 10
	MsgType_APP_MSG_CONSENSUS_PACKET MsgType = 11
	MsgType_APP_MSG_P2PRTSYN         MsgType = 12
	MsgType_APP_MSG_P2PRTSYNACK      MsgType = 13
	MsgType_APP_MSG_GOSSIP           MsgType = 14
	MsgType_APP_MSG_GOSSIP_PULL      MsgType = 15
	MsgType_APP_MSG_DKGSIJ           MsgType = 16
	MsgType_APP_MSG_DKGNLQUAL        MsgType = 17
	MsgType_APP_MSG_DKGLQUAL         MsgType = 18
	MsgType_APP_MSG_UNDEFINED        MsgType = 19
)

var MsgType_name = map[int32]string{
	0:  "APP_MSG_TRN",
	1:  "APP_MSG_BLK",
	2:  "APP_MSG_SIGNPRE",
	3:  "APP_MSG_BLKF",
	4:  "APP_MSG_SIGNBLKF",
	5:  "APP_MSG_BLKS",
	6:  "APP_MSG_REQSYN",
	7:  "APP_MSG_REQSYNSOLO",
	8:  "APP_MSG_BLKSYN",
	9:  "APP_MSG_TIMEOUT",
	10: "APP_MSG_SHARDING_PACKET",
	11: "APP_MSG_CONSENSUS_PACKET",
	12: "APP_MSG_P2PRTSYN",
	13: "APP_MSG_P2PRTSYNACK",
	14: "APP_MSG_GOSSIP",
	15: "APP_MSG_GOSSIP_PULL",
	16: "APP_MSG_DKGSIJ",
	17: "APP_MSG_DKGNLQUAL",
	18: "APP_MSG_DKGLQUAL",
	19: "APP_MSG_UNDEFINED",
}
var MsgType_value = map[string]int32{
	"APP_MSG_TRN":              0,
	"APP_MSG_BLK":              1,
	"APP_MSG_SIGNPRE":          2,
	"APP_MSG_BLKF":             3,
	"APP_MSG_SIGNBLKF":         4,
	"APP_MSG_BLKS":             5,
	"APP_MSG_REQSYN":           6,
	"APP_MSG_REQSYNSOLO":       7,
	"APP_MSG_BLKSYN":           8,
	"APP_MSG_TIMEOUT":          9,
	"APP_MSG_SHARDING_PACKET":  10,
	"APP_MSG_CONSENSUS_PACKET": 11,
	"APP_MSG_P2PRTSYN":         12,
	"APP_MSG_P2PRTSYNACK":      13,
	"APP_MSG_GOSSIP":           14,
	"APP_MSG_GOSSIP_PULL":      15,
	"APP_MSG_DKGSIJ":           16,
	"APP_MSG_DKGNLQUAL":        17,
	"APP_MSG_DKGLQUAL":         18,
	"APP_MSG_UNDEFINED":        19,
}

func (x MsgType) String() string {
	return proto.EnumName(MsgType_name, int32(x))
}
func (MsgType) EnumDescriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

type Message struct {
	ChainId uint32  `protobuf:"varint,1,opt,name=chainId,proto3" json:"chainId,omitempty"`
	Type    MsgType `protobuf:"varint,2,opt,name=type,proto3,enum=pb.MsgType" json:"type,omitempty"`
	Nonce   uint64  `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Data    []byte  `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

func (m *Message) GetChainId() uint32 {
	if m != nil {
		return m.ChainId
	}
	return 0
}

func (m *Message) GetType() MsgType {
	if m != nil {
		return m.Type
	}
	return MsgType_APP_MSG_TRN
}

func (m *Message) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *Message) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*Message)(nil), "pb.Message")
	proto.RegisterEnum("pb.MsgType", MsgType_name, MsgType_value)
}
func (m *Message) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Message) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.ChainId != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.ChainId))
	}
	if m.Type != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Type))
	}
	if m.Nonce != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Nonce))
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintMessage(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func encodeVarintMessage(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Message) Size() (n int) {
	var l int
	_ = l
	if m.ChainId != 0 {
		n += 1 + sovMessage(uint64(m.ChainId))
	}
	if m.Type != 0 {
		n += 1 + sovMessage(uint64(m.Type))
	}
	if m.Nonce != 0 {
		n += 1 + sovMessage(uint64(m.Nonce))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovMessage(uint64(l))
	}
	return n
}

func sovMessage(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMessage(x uint64) (n int) {
	return sovMessage(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Message) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Message: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Message: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainId", wireType)
			}
			m.ChainId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ChainId |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (MsgType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nonce", wireType)
			}
			m.Nonce = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nonce |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMessage(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthMessage
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMessage
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipMessage(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthMessage = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMessage   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("message.proto", fileDescriptorMessage) }

var fileDescriptorMessage = []byte{
	// 379 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x92, 0xcf, 0x72, 0x94, 0x40,
	0x10, 0xc6, 0x33, 0xbb, 0x24, 0x68, 0xef, 0xbf, 0x4e, 0x6f, 0x34, 0x54, 0x69, 0x21, 0xe5, 0x89,
	0xf2, 0xb0, 0x87, 0xf8, 0x04, 0x64, 0x21, 0x88, 0xc0, 0x30, 0x99, 0x81, 0x43, 0x4e, 0x14, 0x9b,
	0x50, 0xd1, 0x83, 0x2c, 0xe5, 0xee, 0x25, 0x6f, 0xe2, 0xd3, 0x78, 0xf6, 0xe8, 0x23, 0x58, 0xeb,
	0x8b, 0x58, 0x01, 0xb1, 0xc6, 0xdc, 0xa6, 0x7f, 0xdf, 0xaf, 0xab, 0xbf, 0xc3, 0xc0, 0xec, 0x4b,
	0xbd, 0xdb, 0x55, 0xf7, 0xf5, 0xaa, 0xfd, 0xba, 0xdd, 0x6f, 0x69, 0xd4, 0x6e, 0xde, 0x36, 0x60,
	0xa6, 0x3d, 0x24, 0x0b, 0xcc, 0xdb, 0x4f, 0xd5, 0xe7, 0x26, 0xba, 0xb3, 0x98, 0xc3, 0xdc, 0x99,
	0x1c, 0x46, 0x7a, 0x03, 0xc6, 0xfe, 0xa1, 0xad, 0xad, 0x91, 0xc3, 0xdc, 0xf9, 0xc5, 0x64, 0xd5,
	0x6e, 0x56, 0xe9, 0xee, 0x3e, 0x7f, 0x68, 0x6b, 0xd9, 0x05, 0x74, 0x06, 0xc7, 0xcd, 0xb6, 0xb9,
	0xad, 0xad, 0xb1, 0xc3, 0x5c, 0x43, 0xf6, 0x03, 0x11, 0x18, 0x77, 0xd5, 0xbe, 0xb2, 0x0c, 0x87,
	0xb9, 0x53, 0xd9, 0xbd, 0xdf, 0x7d, 0x1f, 0x83, 0xf9, 0x77, 0x97, 0x16, 0x30, 0xf1, 0x84, 0x28,
	0x53, 0x15, 0x96, 0xb9, 0xe4, 0x78, 0xa4, 0x83, 0xcb, 0x24, 0x46, 0x46, 0x4b, 0x58, 0x0c, 0x40,
	0x45, 0x21, 0x17, 0x32, 0xc0, 0x11, 0x21, 0x4c, 0x35, 0xeb, 0x0a, 0xc7, 0x74, 0x06, 0xa8, 0x6b,
	0x1d, 0x35, 0x9e, 0x78, 0x0a, 0x8f, 0x89, 0x60, 0x3e, 0x10, 0x19, 0x5c, 0xab, 0x1b, 0x8e, 0x27,
	0xf4, 0x12, 0xe8, 0x7f, 0xa6, 0xb2, 0x24, 0x43, 0x53, 0x77, 0x1f, 0xb7, 0x6f, 0x38, 0x3e, 0xd3,
	0xeb, 0xe4, 0x51, 0x1a, 0x64, 0x45, 0x8e, 0xcf, 0xe9, 0x15, 0x9c, 0xff, 0x3b, 0xfe, 0xc1, 0x93,
	0x7e, 0xc4, 0xc3, 0x52, 0x78, 0xeb, 0x38, 0xc8, 0x11, 0xe8, 0x35, 0x58, 0x43, 0xb8, 0xce, 0xb8,
	0x0a, 0xb8, 0x2a, 0xd4, 0x90, 0x4e, 0xf4, 0xde, 0xe2, 0x42, 0xc8, 0xfc, 0xf1, 0xca, 0x94, 0xce,
	0x61, 0xf9, 0x94, 0x7a, 0xeb, 0x18, 0x67, 0x7a, 0xa5, 0x30, 0x53, 0x2a, 0x12, 0x38, 0xd7, 0xe5,
	0x9e, 0x95, 0xa2, 0x48, 0x12, 0x5c, 0xe8, 0xb2, 0x1f, 0x87, 0x2a, 0xfa, 0x88, 0x48, 0x2f, 0xe0,
	0x54, 0x63, 0x3c, 0xb9, 0x2e, 0xbc, 0x04, 0x4f, 0xf5, 0x1a, 0x7e, 0x1c, 0xf6, 0x94, 0x74, 0xb9,
	0xe0, 0x7e, 0x70, 0x15, 0xf1, 0xc0, 0xc7, 0xe5, 0x25, 0xfe, 0x38, 0xd8, 0xec, 0xe7, 0xc1, 0x66,
	0xbf, 0x0e, 0x36, 0xfb, 0xf6, 0xdb, 0x3e, 0xda, 0x9c, 0x74, 0xbf, 0xe9, 0xfd, 0x9f, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x0e, 0x83, 0x3e, 0xf9, 0x5e, 0x02, 0x00, 0x00,
}
