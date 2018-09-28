// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: announcement.proto

/*
	Package pb is a generated protocol buffer package.

	It is generated from these files:
		announcement.proto

	It has these top-level messages:
		Announcement
		Proof
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

type Announcement struct {
	PublicKey     []byte `protobuf:"bytes,1,opt,name=PublicKey,proto3" json:"PublicKey,omitempty"`
	TotalStorage  uint64 `protobuf:"varint,2,opt,name=TotalStorage,proto3" json:"TotalStorage,omitempty"`
	StartAt       uint64 `protobuf:"varint,3,opt,name=StartAt,proto3" json:"StartAt,omitempty"`
	Collateral    []byte `protobuf:"bytes,4,opt,name=Collateral,proto3" json:"Collateral,omitempty"`
	MaxCollateral []byte `protobuf:"bytes,5,opt,name=MaxCollateral,proto3" json:"MaxCollateral,omitempty"`
	AccountName   string `protobuf:"bytes,6,opt,name=AccountName,proto3" json:"AccountName,omitempty"`
}

func (m *Announcement) Reset()                    { *m = Announcement{} }
func (m *Announcement) String() string            { return proto.CompactTextString(m) }
func (*Announcement) ProtoMessage()               {}
func (*Announcement) Descriptor() ([]byte, []int) { return fileDescriptorAnnouncement, []int{0} }

func (m *Announcement) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *Announcement) GetTotalStorage() uint64 {
	if m != nil {
		return m.TotalStorage
	}
	return 0
}

func (m *Announcement) GetStartAt() uint64 {
	if m != nil {
		return m.StartAt
	}
	return 0
}

func (m *Announcement) GetCollateral() []byte {
	if m != nil {
		return m.Collateral
	}
	return nil
}

func (m *Announcement) GetMaxCollateral() []byte {
	if m != nil {
		return m.MaxCollateral
	}
	return nil
}

func (m *Announcement) GetAccountName() string {
	if m != nil {
		return m.AccountName
	}
	return ""
}

type Proof struct {
	PublicKey    []byte `protobuf:"bytes,1,opt,name=PublicKey,proto3" json:"PublicKey,omitempty"`
	RepoSize     uint64 `protobuf:"varint,2,opt,name=RepoSize,proto3" json:"RepoSize,omitempty"`
	Cid          string `protobuf:"bytes,3,opt,name=Cid,proto3" json:"Cid,omitempty"`
	SegmentIndex uint64 `protobuf:"varint,4,opt,name=SegmentIndex,proto3" json:"SegmentIndex,omitempty"`
	Segment      []byte `protobuf:"bytes,5,opt,name=Segment,proto3" json:"Segment,omitempty"`
	HashSet      []byte `protobuf:"bytes,6,opt,name=HashSet,proto3" json:"HashSet,omitempty"`
	AtHeight     uint64 `protobuf:"varint,7,opt,name=AtHeight,proto3" json:"AtHeight,omitempty"`
	AccountName  string `protobuf:"bytes,8,opt,name=AccountName,proto3" json:"AccountName,omitempty"`
}

func (m *Proof) Reset()                    { *m = Proof{} }
func (m *Proof) String() string            { return proto.CompactTextString(m) }
func (*Proof) ProtoMessage()               {}
func (*Proof) Descriptor() ([]byte, []int) { return fileDescriptorAnnouncement, []int{1} }

func (m *Proof) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *Proof) GetRepoSize() uint64 {
	if m != nil {
		return m.RepoSize
	}
	return 0
}

func (m *Proof) GetCid() string {
	if m != nil {
		return m.Cid
	}
	return ""
}

func (m *Proof) GetSegmentIndex() uint64 {
	if m != nil {
		return m.SegmentIndex
	}
	return 0
}

func (m *Proof) GetSegment() []byte {
	if m != nil {
		return m.Segment
	}
	return nil
}

func (m *Proof) GetHashSet() []byte {
	if m != nil {
		return m.HashSet
	}
	return nil
}

func (m *Proof) GetAtHeight() uint64 {
	if m != nil {
		return m.AtHeight
	}
	return 0
}

func (m *Proof) GetAccountName() string {
	if m != nil {
		return m.AccountName
	}
	return ""
}

func init() {
	proto.RegisterType((*Announcement)(nil), "pb.Announcement")
	proto.RegisterType((*Proof)(nil), "pb.Proof")
}
func (m *Announcement) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Announcement) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PublicKey) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.PublicKey)))
		i += copy(dAtA[i:], m.PublicKey)
	}
	if m.TotalStorage != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(m.TotalStorage))
	}
	if m.StartAt != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(m.StartAt))
	}
	if len(m.Collateral) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.Collateral)))
		i += copy(dAtA[i:], m.Collateral)
	}
	if len(m.MaxCollateral) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.MaxCollateral)))
		i += copy(dAtA[i:], m.MaxCollateral)
	}
	if len(m.AccountName) > 0 {
		dAtA[i] = 0x32
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.AccountName)))
		i += copy(dAtA[i:], m.AccountName)
	}
	return i, nil
}

func (m *Proof) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Proof) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PublicKey) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.PublicKey)))
		i += copy(dAtA[i:], m.PublicKey)
	}
	if m.RepoSize != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(m.RepoSize))
	}
	if len(m.Cid) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.Cid)))
		i += copy(dAtA[i:], m.Cid)
	}
	if m.SegmentIndex != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(m.SegmentIndex))
	}
	if len(m.Segment) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.Segment)))
		i += copy(dAtA[i:], m.Segment)
	}
	if len(m.HashSet) > 0 {
		dAtA[i] = 0x32
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.HashSet)))
		i += copy(dAtA[i:], m.HashSet)
	}
	if m.AtHeight != 0 {
		dAtA[i] = 0x38
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(m.AtHeight))
	}
	if len(m.AccountName) > 0 {
		dAtA[i] = 0x42
		i++
		i = encodeVarintAnnouncement(dAtA, i, uint64(len(m.AccountName)))
		i += copy(dAtA[i:], m.AccountName)
	}
	return i, nil
}

func encodeVarintAnnouncement(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Announcement) Size() (n int) {
	var l int
	_ = l
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	if m.TotalStorage != 0 {
		n += 1 + sovAnnouncement(uint64(m.TotalStorage))
	}
	if m.StartAt != 0 {
		n += 1 + sovAnnouncement(uint64(m.StartAt))
	}
	l = len(m.Collateral)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	l = len(m.MaxCollateral)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	l = len(m.AccountName)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	return n
}

func (m *Proof) Size() (n int) {
	var l int
	_ = l
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	if m.RepoSize != 0 {
		n += 1 + sovAnnouncement(uint64(m.RepoSize))
	}
	l = len(m.Cid)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	if m.SegmentIndex != 0 {
		n += 1 + sovAnnouncement(uint64(m.SegmentIndex))
	}
	l = len(m.Segment)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	l = len(m.HashSet)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	if m.AtHeight != 0 {
		n += 1 + sovAnnouncement(uint64(m.AtHeight))
	}
	l = len(m.AccountName)
	if l > 0 {
		n += 1 + l + sovAnnouncement(uint64(l))
	}
	return n
}

func sovAnnouncement(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozAnnouncement(x uint64) (n int) {
	return sovAnnouncement(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Announcement) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAnnouncement
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
			return fmt.Errorf("proto: Announcement: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Announcement: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
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
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKey = append(m.PublicKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PublicKey == nil {
				m.PublicKey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalStorage", wireType)
			}
			m.TotalStorage = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalStorage |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartAt", wireType)
			}
			m.StartAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StartAt |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Collateral", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
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
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Collateral = append(m.Collateral[:0], dAtA[iNdEx:postIndex]...)
			if m.Collateral == nil {
				m.Collateral = []byte{}
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxCollateral", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
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
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MaxCollateral = append(m.MaxCollateral[:0], dAtA[iNdEx:postIndex]...)
			if m.MaxCollateral == nil {
				m.MaxCollateral = []byte{}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAnnouncement(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAnnouncement
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
func (m *Proof) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAnnouncement
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
			return fmt.Errorf("proto: Proof: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Proof: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
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
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKey = append(m.PublicKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PublicKey == nil {
				m.PublicKey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RepoSize", wireType)
			}
			m.RepoSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RepoSize |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cid", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cid = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SegmentIndex", wireType)
			}
			m.SegmentIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SegmentIndex |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Segment", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
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
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Segment = append(m.Segment[:0], dAtA[iNdEx:postIndex]...)
			if m.Segment == nil {
				m.Segment = []byte{}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HashSet", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
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
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.HashSet = append(m.HashSet[:0], dAtA[iNdEx:postIndex]...)
			if m.HashSet == nil {
				m.HashSet = []byte{}
			}
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AtHeight", wireType)
			}
			m.AtHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AtHeight |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnnouncement
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnnouncement
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAnnouncement(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAnnouncement
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
func skipAnnouncement(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAnnouncement
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
					return 0, ErrIntOverflowAnnouncement
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
					return 0, ErrIntOverflowAnnouncement
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
				return 0, ErrInvalidLengthAnnouncement
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowAnnouncement
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
				next, err := skipAnnouncement(dAtA[start:])
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
	ErrInvalidLengthAnnouncement = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAnnouncement   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("announcement.proto", fileDescriptorAnnouncement) }

var fileDescriptorAnnouncement = []byte{
	// 298 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0x4f, 0x4a, 0xc3, 0x40,
	0x14, 0xc6, 0x9d, 0xfe, 0xef, 0x33, 0x42, 0x99, 0xd5, 0x20, 0x12, 0x42, 0x71, 0x91, 0x95, 0x1b,
	0x4f, 0x10, 0xbb, 0xa9, 0x88, 0x52, 0x26, 0x5e, 0x60, 0x92, 0x3e, 0xd3, 0x40, 0x3a, 0x13, 0xe2,
	0x0b, 0x54, 0x4f, 0xe2, 0x91, 0x5c, 0x8a, 0x27, 0x90, 0xb8, 0xf2, 0x16, 0x92, 0xa9, 0xd1, 0x44,
	0x17, 0xee, 0xe6, 0xfb, 0xbd, 0xc7, 0xf0, 0xfd, 0x66, 0x80, 0x2b, 0xad, 0x4d, 0xa9, 0x63, 0xdc,
	0xa2, 0xa6, 0xb3, 0xbc, 0x30, 0x64, 0x78, 0x2f, 0x8f, 0xe6, 0xaf, 0x0c, 0x9c, 0xa0, 0x35, 0xe2,
	0x27, 0x30, 0x5d, 0x95, 0x51, 0x96, 0xc6, 0x57, 0xf8, 0x20, 0x98, 0xc7, 0x7c, 0x47, 0xfe, 0x00,
	0x3e, 0x07, 0xe7, 0xd6, 0x90, 0xca, 0x42, 0x32, 0x85, 0x4a, 0x50, 0xf4, 0x3c, 0xe6, 0x0f, 0x64,
	0x87, 0x71, 0x01, 0xe3, 0x90, 0x54, 0x41, 0x01, 0x89, 0xbe, 0x1d, 0x37, 0x91, 0xbb, 0x00, 0x0b,
	0x93, 0x65, 0x8a, 0xb0, 0x50, 0x99, 0x18, 0xd8, 0xcb, 0x5b, 0x84, 0x9f, 0xc2, 0xd1, 0xb5, 0xda,
	0xb5, 0x56, 0x86, 0x76, 0xa5, 0x0b, 0xb9, 0x07, 0x87, 0x41, 0x1c, 0x9b, 0x52, 0xd3, 0x8d, 0xda,
	0xa2, 0x18, 0x79, 0xcc, 0x9f, 0xca, 0x36, 0x9a, 0x7f, 0x30, 0x18, 0xae, 0x0a, 0x63, 0xee, 0xfe,
	0xb1, 0x39, 0x86, 0x89, 0xc4, 0xdc, 0x84, 0xe9, 0x63, 0x63, 0xf2, 0x9d, 0xf9, 0x0c, 0xfa, 0x8b,
	0x74, 0x6d, 0x0d, 0xa6, 0xb2, 0x3e, 0xd6, 0xee, 0x21, 0x26, 0xf5, 0x23, 0x5d, 0xea, 0x35, 0xee,
	0x6c, 0xff, 0x81, 0xec, 0x30, 0xeb, 0xbe, 0xcf, 0x5f, 0xdd, 0x9b, 0x58, 0x4f, 0x96, 0xea, 0x7e,
	0x13, 0x22, 0xd9, 0xc6, 0x8e, 0x6c, 0x62, 0xdd, 0x22, 0xa0, 0x25, 0xa6, 0xc9, 0x86, 0xc4, 0x78,
	0xdf, 0xa2, 0xc9, 0xbf, 0x5d, 0x27, 0x7f, 0x5c, 0x2f, 0x66, 0xcf, 0x95, 0xcb, 0x5e, 0x2a, 0x97,
	0xbd, 0x55, 0x2e, 0x7b, 0x7a, 0x77, 0x0f, 0xa2, 0x91, 0xfd, 0xdd, 0xf3, 0xcf, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x05, 0xf8, 0x29, 0x6c, 0xf3, 0x01, 0x00, 0x00,
}