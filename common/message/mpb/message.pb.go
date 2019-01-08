// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: message.proto

/*
	Package mpb is a generated protocol buffer package.

	It is generated from these files:
		message.proto

	It has these top-level messages:
		Message
*/
package mpb

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
type Identify int32

const (
	Identify_APP_MSG_HASH                 Identify = 0
	Identify_APP_MSG_ADDRESS              Identify = 1
	Identify_APP_MSG_TRANSACTION          Identify = 2
	Identify_APP_MSG_TRANSACTION_DEPLOY   Identify = 3
	Identify_APP_MSG_TRANSACTION_INVOKE   Identify = 4
	Identify_APP_MSG_TRANSACTION_TRANSFER Identify = 5
	Identify_APP_MSG_HEADER               Identify = 6
	Identify_APP_MSG_BLOCK                Identify = 7
	Identify_APP_MSG_SHARD_BLOCK          Identify = 8
	Identify_APP_MSG_MINOR_BLOCK          Identify = 9
	Identify_APP_MSG_CM_BLOCK             Identify = 10
	Identify_APP_MSG_FINAL_BLOCK          Identify = 11
	Identify_APP_MSG_VC_BLOCK             Identify = 12
	Identify_APP_MSG_SIGNPRE              Identify = 13
	Identify_APP_MSG_TIMEOUT              Identify = 14
	Identify_APP_MSG_SHARDING_PACKET      Identify = 15
	Identify_APP_MSG_CONSENSUS_PACKET     Identify = 16
	Identify_APP_MSG_P2PRTSYN             Identify = 17
	Identify_APP_MSG_P2PRTSYNACK          Identify = 18
	Identify_APP_MSG_GOSSIP               Identify = 19
	Identify_APP_MSG_GOSSIP_PULL          Identify = 20
	Identify_APP_MSG_DKGSIJ               Identify = 21
	Identify_APP_MSG_DKGNLQUAL            Identify = 22
	Identify_APP_MSG_DKGLQUAL             Identify = 23
	Identify_APP_MSG_SYNC_REQUEST         Identify = 24
	Identify_APP_MSG_SYNC_RESPONSE        Identify = 25
	Identify_APP_MSG_STRING               Identify = 26
	Identify_APP_MSG_EBALLSCAN_HEIGHT     Identify = 27
	Identify_APP_MSG_BLOCK_REQUEST        Identify = 28
	Identify_APP_MSG_BLOCK_RESPONSE       Identify = 29
	Identify_APP_MSG_STATE_OBJECT         Identify = 30
	Identify_APP_MSG_TRANSACTION_RECEIPT  Identify = 31
	Identify_APP_MSG_ACCOUNT_PERMISSION   Identify = 32
	Identify_APP_MSG_ACCOUNT_RESOURCE     Identify = 33
	Identify_APP_MSG_UNDEFINED            Identify = 34
)

var Identify_name = map[int32]string{
	0:  "APP_MSG_HASH",
	1:  "APP_MSG_ADDRESS",
	2:  "APP_MSG_TRANSACTION",
	3:  "APP_MSG_TRANSACTION_DEPLOY",
	4:  "APP_MSG_TRANSACTION_INVOKE",
	5:  "APP_MSG_TRANSACTION_TRANSFER",
	6:  "APP_MSG_HEADER",
	7:  "APP_MSG_BLOCK",
	8:  "APP_MSG_SHARD_BLOCK",
	9:  "APP_MSG_MINOR_BLOCK",
	10: "APP_MSG_CM_BLOCK",
	11: "APP_MSG_FINAL_BLOCK",
	12: "APP_MSG_VC_BLOCK",
	13: "APP_MSG_SIGNPRE",
	14: "APP_MSG_TIMEOUT",
	15: "APP_MSG_SHARDING_PACKET",
	16: "APP_MSG_CONSENSUS_PACKET",
	17: "APP_MSG_P2PRTSYN",
	18: "APP_MSG_P2PRTSYNACK",
	19: "APP_MSG_GOSSIP",
	20: "APP_MSG_GOSSIP_PULL",
	21: "APP_MSG_DKGSIJ",
	22: "APP_MSG_DKGNLQUAL",
	23: "APP_MSG_DKGLQUAL",
	24: "APP_MSG_SYNC_REQUEST",
	25: "APP_MSG_SYNC_RESPONSE",
	26: "APP_MSG_STRING",
	27: "APP_MSG_EBALLSCAN_HEIGHT",
	28: "APP_MSG_BLOCK_REQUEST",
	29: "APP_MSG_BLOCK_RESPONSE",
	30: "APP_MSG_STATE_OBJECT",
	31: "APP_MSG_TRANSACTION_RECEIPT",
	32: "APP_MSG_ACCOUNT_PERMISSION",
	33: "APP_MSG_ACCOUNT_RESOURCE",
	34: "APP_MSG_UNDEFINED",
}
var Identify_value = map[string]int32{
	"APP_MSG_HASH":                 0,
	"APP_MSG_ADDRESS":              1,
	"APP_MSG_TRANSACTION":          2,
	"APP_MSG_TRANSACTION_DEPLOY":   3,
	"APP_MSG_TRANSACTION_INVOKE":   4,
	"APP_MSG_TRANSACTION_TRANSFER": 5,
	"APP_MSG_HEADER":               6,
	"APP_MSG_BLOCK":                7,
	"APP_MSG_SHARD_BLOCK":          8,
	"APP_MSG_MINOR_BLOCK":          9,
	"APP_MSG_CM_BLOCK":             10,
	"APP_MSG_FINAL_BLOCK":          11,
	"APP_MSG_VC_BLOCK":             12,
	"APP_MSG_SIGNPRE":              13,
	"APP_MSG_TIMEOUT":              14,
	"APP_MSG_SHARDING_PACKET":      15,
	"APP_MSG_CONSENSUS_PACKET":     16,
	"APP_MSG_P2PRTSYN":             17,
	"APP_MSG_P2PRTSYNACK":          18,
	"APP_MSG_GOSSIP":               19,
	"APP_MSG_GOSSIP_PULL":          20,
	"APP_MSG_DKGSIJ":               21,
	"APP_MSG_DKGNLQUAL":            22,
	"APP_MSG_DKGLQUAL":             23,
	"APP_MSG_SYNC_REQUEST":         24,
	"APP_MSG_SYNC_RESPONSE":        25,
	"APP_MSG_STRING":               26,
	"APP_MSG_EBALLSCAN_HEIGHT":     27,
	"APP_MSG_BLOCK_REQUEST":        28,
	"APP_MSG_BLOCK_RESPONSE":       29,
	"APP_MSG_STATE_OBJECT":         30,
	"APP_MSG_TRANSACTION_RECEIPT":  31,
	"APP_MSG_ACCOUNT_PERMISSION":   32,
	"APP_MSG_ACCOUNT_RESOURCE":     33,
	"APP_MSG_UNDEFINED":            34,
}

func (x Identify) String() string {
	return proto.EnumName(Identify_name, int32(x))
}
func (Identify) EnumDescriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

type Message struct {
	Nonce    uint64   `protobuf:"varint,1,opt,name=Nonce,proto3" json:"Nonce,omitempty"`
	Identify Identify `protobuf:"varint,2,opt,name=Identify,proto3,enum=mpb.Identify" json:"Identify,omitempty"`
	Payload  []byte   `protobuf:"bytes,4,opt,name=Payload,proto3" json:"Payload,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

func (m *Message) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *Message) GetIdentify() Identify {
	if m != nil {
		return m.Identify
	}
	return Identify_APP_MSG_HASH
}

func (m *Message) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func init() {
	proto.RegisterType((*Message)(nil), "mpb.Message")
	proto.RegisterEnum("mpb.Identify", Identify_name, Identify_value)
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
	if m.Nonce != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Nonce))
	}
	if m.Identify != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.Identify))
	}
	if len(m.Payload) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintMessage(dAtA, i, uint64(len(m.Payload)))
		i += copy(dAtA[i:], m.Payload)
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
	if m.Nonce != 0 {
		n += 1 + sovMessage(uint64(m.Nonce))
	}
	if m.Identify != 0 {
		n += 1 + sovMessage(uint64(m.Identify))
	}
	l = len(m.Payload)
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
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Identify", wireType)
			}
			m.Identify = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Identify |= (Identify(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Payload", wireType)
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
			m.Payload = append(m.Payload[:0], dAtA[iNdEx:postIndex]...)
			if m.Payload == nil {
				m.Payload = []byte{}
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
	// 539 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x94, 0xdf, 0x6e, 0xda, 0x30,
	0x14, 0x87, 0x9b, 0xfe, 0x9f, 0x57, 0xda, 0x53, 0x03, 0x25, 0x05, 0x96, 0xb2, 0x5e, 0xb1, 0x5d,
	0x70, 0xd1, 0x3d, 0x81, 0x71, 0x4c, 0x70, 0x93, 0x38, 0xae, 0xed, 0x54, 0xe2, 0x2a, 0x82, 0xc1,
	0xa6, 0x49, 0xa3, 0xa0, 0xb5, 0x37, 0x7d, 0x93, 0x3d, 0xd2, 0x2e, 0xf7, 0x08, 0x13, 0x7b, 0x8b,
	0x5d, 0x4d, 0x85, 0x84, 0xb9, 0x15, 0xbb, 0x8b, 0xbf, 0xef, 0xc8, 0xe7, 0xe7, 0xa3, 0xa3, 0xa0,
	0xd2, 0x74, 0x72, 0x7f, 0x3f, 0xfc, 0x3c, 0xe9, 0xcc, 0xbf, 0xcd, 0x1e, 0x66, 0x78, 0x67, 0x3a,
	0x1f, 0x5d, 0x8e, 0xd1, 0x41, 0xbc, 0xa2, 0xb8, 0x82, 0xf6, 0xc4, 0xec, 0xee, 0xe3, 0xc4, 0x75,
	0x5a, 0x4e, 0x7b, 0x57, 0xad, 0x0e, 0xf8, 0x1d, 0x3a, 0xe4, 0xe3, 0xc9, 0xdd, 0xc3, 0x97, 0x4f,
	0x8f, 0xee, 0x76, 0xcb, 0x69, 0x1f, 0x5f, 0x95, 0x3a, 0xd3, 0xf9, 0xa8, 0x53, 0x40, 0xb5, 0xd6,
	0xd8, 0x45, 0x07, 0x72, 0xf8, 0xf8, 0x75, 0x36, 0x1c, 0xbb, 0xbb, 0x2d, 0xa7, 0x7d, 0xa4, 0x8a,
	0xe3, 0xfb, 0x3f, 0xfb, 0xff, 0x6e, 0xc1, 0x80, 0x8e, 0x88, 0x94, 0x59, 0xac, 0x83, 0xac, 0x4f,
	0x74, 0x1f, 0xb6, 0x70, 0x19, 0x9d, 0x14, 0x84, 0xf8, 0xbe, 0x62, 0x5a, 0x83, 0x83, 0x6b, 0xa8,
	0x5c, 0x40, 0xa3, 0x88, 0xd0, 0x84, 0x1a, 0x9e, 0x08, 0xd8, 0xc6, 0x1e, 0xaa, 0x6f, 0x10, 0x99,
	0xcf, 0x64, 0x94, 0x0c, 0x60, 0xe7, 0x7f, 0x9e, 0x8b, 0xdb, 0x24, 0x64, 0xb0, 0x8b, 0x5b, 0xa8,
	0xb9, 0xc9, 0x2f, 0xbf, 0x7b, 0x4c, 0xc1, 0x1e, 0xc6, 0xe8, 0x78, 0x9d, 0x90, 0x11, 0x9f, 0x29,
	0xd8, 0xc7, 0xa7, 0xa8, 0x54, 0xb0, 0x6e, 0x94, 0xd0, 0x10, 0x0e, 0xec, 0x84, 0xba, 0x4f, 0x94,
	0x9f, 0x8b, 0x43, 0x5b, 0xc4, 0x5c, 0x24, 0x2a, 0x17, 0xaf, 0x70, 0x05, 0x41, 0x21, 0x68, 0x9c,
	0x53, 0x64, 0x97, 0xf7, 0xb8, 0x20, 0x51, 0x2e, 0x5e, 0xdb, 0xe5, 0xb7, 0x34, 0xa7, 0x47, 0xf6,
	0xb4, 0x34, 0x0f, 0x84, 0x54, 0x0c, 0x4a, 0x36, 0x34, 0x3c, 0x66, 0x49, 0x6a, 0xe0, 0x18, 0x37,
	0x50, 0xed, 0x59, 0x40, 0x2e, 0x82, 0x4c, 0x12, 0x1a, 0x32, 0x03, 0x27, 0xb8, 0x89, 0xdc, 0x75,
	0x96, 0x44, 0x68, 0x26, 0x74, 0xaa, 0x0b, 0x0b, 0x76, 0x6b, 0x79, 0x25, 0x95, 0xd1, 0x03, 0x01,
	0xa7, 0x76, 0xd2, 0x82, 0x12, 0x1a, 0x02, 0xb6, 0x27, 0x16, 0x24, 0x5a, 0x73, 0x09, 0x65, 0xbb,
	0x78, 0xc5, 0x32, 0x99, 0x46, 0x11, 0x54, 0xec, 0x62, 0x3f, 0x0c, 0x34, 0xbf, 0x86, 0x2a, 0xae,
	0xa2, 0x53, 0x8b, 0x89, 0xe8, 0x26, 0x25, 0x11, 0x9c, 0xd9, 0x31, 0xfc, 0x30, 0x58, 0xd1, 0x1a,
	0x76, 0x51, 0x65, 0xfd, 0xae, 0x81, 0xa0, 0x99, 0x62, 0x37, 0x29, 0xd3, 0x06, 0x5c, 0x7c, 0x8e,
	0xaa, 0x2f, 0x8c, 0x96, 0x4f, 0x8f, 0x83, 0x73, 0xbb, 0xab, 0x36, 0x8a, 0x8b, 0x00, 0xea, 0xf6,
	0x0c, 0x58, 0x97, 0x44, 0x91, 0xa6, 0x44, 0x64, 0x7d, 0xc6, 0x83, 0xbe, 0x81, 0x86, 0x7d, 0xd9,
	0x72, 0xf6, 0xeb, 0x3e, 0x4d, 0x5c, 0x47, 0x67, 0x2f, 0x55, 0xde, 0xe8, 0xcd, 0xb3, 0x74, 0x86,
	0x18, 0x96, 0x25, 0xdd, 0x6b, 0x46, 0x0d, 0x78, 0xf8, 0x02, 0x35, 0x36, 0x6d, 0x9e, 0x62, 0x94,
	0x71, 0x69, 0xe0, 0xc2, 0x5e, 0x5d, 0x42, 0x69, 0x92, 0x0a, 0x93, 0x49, 0xa6, 0x62, 0xae, 0xf5,
	0xd3, 0xea, 0xb7, 0xec, 0xbc, 0x85, 0x57, 0x4c, 0x27, 0xa9, 0xa2, 0x0c, 0xde, 0xda, 0x33, 0x4c,
	0x85, 0xcf, 0x7a, 0x5c, 0x30, 0x1f, 0x2e, 0xbb, 0xf0, 0x63, 0xe1, 0x39, 0x3f, 0x17, 0x9e, 0xf3,
	0x6b, 0xe1, 0x39, 0xdf, 0x7f, 0x7b, 0x5b, 0xa3, 0xfd, 0xe5, 0x0f, 0xe0, 0xc3, 0xdf, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x5e, 0x75, 0x72, 0xb2, 0x11, 0x04, 0x00, 0x00,
}
