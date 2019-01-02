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
	Identify_APP_MSG_HASH             Identify = 0
	Identify_APP_MSG_ADDRESS          Identify = 1
	Identify_APP_MSG_TRANSACTION      Identify = 2
	Identify_APP_MSG_SHARD_BLOCK      Identify = 3
	Identify_APP_MSG_MINOR_BLOCK      Identify = 4
	Identify_APP_MSG_CM_BLOCK         Identify = 5
	Identify_APP_MSG_FINAL_BLOCK      Identify = 6
	Identify_APP_MSG_VC_BLOCK         Identify = 7
	Identify_APP_MSG_SIGNPRE          Identify = 8
	Identify_APP_MSG_TIMEOUT          Identify = 9
	Identify_APP_MSG_SHARDING_PACKET  Identify = 10
	Identify_APP_MSG_CONSENSUS_PACKET Identify = 11
	Identify_APP_MSG_P2PRTSYN         Identify = 12
	Identify_APP_MSG_P2PRTSYNACK      Identify = 13
	Identify_APP_MSG_GOSSIP           Identify = 14
	Identify_APP_MSG_GOSSIP_PULL      Identify = 15
	Identify_APP_MSG_DKGSIJ           Identify = 16
	Identify_APP_MSG_DKGNLQUAL        Identify = 17
	Identify_APP_MSG_DKGLQUAL         Identify = 18
	Identify_APP_MSG_SYNC_REQUEST     Identify = 19
	Identify_APP_MSG_SYNC_RESPONSE    Identify = 20
	Identify_APP_MSG_STRING           Identify = 21
	Identify_APP_MSG_UNDEFINED        Identify = 22
)

var Identify_name = map[int32]string{
	0:  "APP_MSG_HASH",
	1:  "APP_MSG_ADDRESS",
	2:  "APP_MSG_TRANSACTION",
	3:  "APP_MSG_SHARD_BLOCK",
	4:  "APP_MSG_MINOR_BLOCK",
	5:  "APP_MSG_CM_BLOCK",
	6:  "APP_MSG_FINAL_BLOCK",
	7:  "APP_MSG_VC_BLOCK",
	8:  "APP_MSG_SIGNPRE",
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
	19: "APP_MSG_SYNC_REQUEST",
	20: "APP_MSG_SYNC_RESPONSE",
	21: "APP_MSG_STRING",
	22: "APP_MSG_UNDEFINED",
}
var Identify_value = map[string]int32{
	"APP_MSG_HASH":             0,
	"APP_MSG_ADDRESS":          1,
	"APP_MSG_TRANSACTION":      2,
	"APP_MSG_SHARD_BLOCK":      3,
	"APP_MSG_MINOR_BLOCK":      4,
	"APP_MSG_CM_BLOCK":         5,
	"APP_MSG_FINAL_BLOCK":      6,
	"APP_MSG_VC_BLOCK":         7,
	"APP_MSG_SIGNPRE":          8,
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
	"APP_MSG_SYNC_REQUEST":     19,
	"APP_MSG_SYNC_RESPONSE":    20,
	"APP_MSG_STRING":           21,
	"APP_MSG_UNDEFINED":        22,
}

func (x Identify) String() string {
	return proto.EnumName(Identify_name, int32(x))
}
func (Identify) EnumDescriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

type Message struct {
	Identify Identify `protobuf:"varint,2,opt,name=Identify,proto3,enum=mpb.Identify" json:"Identify,omitempty"`
	Payload  []byte   `protobuf:"bytes,4,opt,name=Payload,proto3" json:"Payload,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{Identify: 0, Payload:  nil,}
}
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

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
	// 395 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x92, 0xc1, 0x8e, 0xd2, 0x40,
	0x18, 0xc7, 0xb7, 0xbb, 0xb8, 0xac, 0x23, 0xb0, 0xdf, 0x0e, 0xe0, 0xd6, 0x68, 0x1a, 0xe2, 0x09,
	0x3d, 0x70, 0xc0, 0x27, 0x18, 0xda, 0xa1, 0x8c, 0x6d, 0xa7, 0xc3, 0x7c, 0x53, 0x13, 0x4e, 0x0d,
	0x04, 0x34, 0x26, 0x22, 0x44, 0xb8, 0xf0, 0x26, 0xde, 0x7d, 0x19, 0x8f, 0x3e, 0x82, 0xc1, 0x17,
	0x31, 0x4a, 0x07, 0xc7, 0x3d, 0xf6, 0xf7, 0xfb, 0xa7, 0xff, 0xff, 0x97, 0x0c, 0x69, 0xae, 0x57,
	0xbb, 0xdd, 0xfc, 0xc3, 0x6a, 0xb0, 0xfd, 0xb2, 0xd9, 0x6f, 0xe8, 0xd5, 0x7a, 0xbb, 0x78, 0x29,
	0x49, 0x3d, 0x3b, 0x51, 0xfa, 0x8a, 0xdc, 0x88, 0xe5, 0xea, 0xf3, 0xfe, 0xe3, 0xfb, 0x83, 0x7f,
	0xd9, 0xf3, 0xfa, 0xad, 0x61, 0x73, 0xb0, 0xde, 0x2e, 0x06, 0x16, 0xea, 0xb3, 0xa6, 0x3e, 0xa9,
	0xab, 0xf9, 0xe1, 0xd3, 0x66, 0xbe, 0xf4, 0x6b, 0x3d, 0xaf, 0xdf, 0xd0, 0xf6, 0xf3, 0xf5, 0xb7,
	0xda, 0xbf, 0xbf, 0x50, 0x20, 0x0d, 0xa6, 0x54, 0x99, 0x61, 0x5c, 0x4e, 0x18, 0x4e, 0xe0, 0x82,
	0xb6, 0xc9, 0xad, 0x25, 0x2c, 0x8a, 0x34, 0x47, 0x04, 0x8f, 0xde, 0x93, 0xb6, 0x85, 0x46, 0x33,
	0x89, 0x2c, 0x34, 0x22, 0x97, 0x70, 0xe9, 0x0a, 0x9c, 0x30, 0x1d, 0x95, 0xa3, 0x34, 0x0f, 0x13,
	0xb8, 0x72, 0x45, 0x26, 0x64, 0xae, 0x2b, 0x51, 0xa3, 0x1d, 0x02, 0x56, 0x84, 0x59, 0x45, 0x1f,
	0xb9, 0xf1, 0xb1, 0x90, 0x2c, 0xad, 0xc4, 0xb5, 0x1b, 0x7f, 0x17, 0x56, 0xb4, 0xee, 0x8e, 0x44,
	0x11, 0x4b, 0xa5, 0x39, 0xdc, 0xb8, 0xd0, 0x88, 0x8c, 0xe7, 0x85, 0x81, 0xc7, 0xf4, 0x39, 0xb9,
	0xff, 0x6f, 0xa0, 0x90, 0x71, 0xa9, 0x58, 0x98, 0x70, 0x03, 0x84, 0xbe, 0x20, 0xfe, 0x79, 0x4b,
	0x2e, 0x91, 0x4b, 0x2c, 0xd0, 0xda, 0x27, 0x6e, 0xb5, 0x1a, 0x2a, 0x6d, 0x70, 0x26, 0xa1, 0xe1,
	0x2e, 0xb5, 0x94, 0x85, 0x09, 0x34, 0x29, 0x25, 0x2d, 0x2b, 0xe2, 0x1c, 0x51, 0x28, 0x68, 0xb9,
	0xe1, 0x13, 0x2b, 0x55, 0x91, 0xa6, 0x70, 0xeb, 0x86, 0xa3, 0x24, 0x46, 0xf1, 0x16, 0x80, 0x76,
	0xc9, 0x9d, 0xc3, 0x64, 0x3a, 0x2d, 0x58, 0x0a, 0x77, 0xee, 0x8c, 0x28, 0x89, 0x4f, 0x94, 0x52,
	0x9f, 0x74, 0xce, 0x77, 0xcd, 0x64, 0x58, 0x6a, 0x3e, 0x2d, 0x38, 0x1a, 0x68, 0xd3, 0x67, 0xa4,
	0xfb, 0xc0, 0xa0, 0xfa, 0x73, 0x1c, 0x74, 0xdc, 0x56, 0x34, 0x5a, 0xc8, 0x18, 0xba, 0x6e, 0x6b,
	0x21, 0x23, 0x3e, 0x16, 0x92, 0x47, 0xf0, 0x74, 0x04, 0xdf, 0x8f, 0x81, 0xf7, 0xe3, 0x18, 0x78,
	0x3f, 0x8f, 0x81, 0xf7, 0xf5, 0x57, 0x70, 0xb1, 0xb8, 0xfe, 0xfb, 0x26, 0xdf, 0xfc, 0x0e, 0x00,
	0x00, 0xff, 0xff, 0x7d, 0x2b, 0x65, 0x6e, 0xa4, 0x02, 0x00, 0x00,
}
