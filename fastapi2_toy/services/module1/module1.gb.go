package module1

import "encoding/binary"
import "github.com/funny/gobuf"
import "github.com/2dgamer/fastapi2"
import "github.com/2dgamer/fastapi2/fastapi2_toy/services"

type MessageID byte

const (
	MsgID_Add MessageID = 0
)

func (_ *Module1) NewRequest(id byte) fastapi2.IMessage {
	switch MessageID(id) {
	case MsgID_Add:
		return &AddReq{}
	}
	return nil
}

func (_ *Module1) NewResponse(id byte) fastapi2.IMessage {
	switch MessageID(id) {
	case MsgID_Add:
		return &AddRsp{}
	}
	return nil
}

func (s *Module1) HandleRequest(session *link.Session, req fastapi2.IMessage) {
	switch MessageID(req.MessageID()) {
	case MsgID_Add:
		session.Send(s.Add(session, req.(*AddReq)))
	default:
		panic("Unhandled Message Type")
	}
}

func (s *Module1) ServiceID() byte {
	return byte(services.ServiceID_Module1)
}

var _ gobuf.Struct = (*AddReq)(nil)

func (s *AddReq) Size() int {
	var size int
	size += gobuf.VarintSize(int64(s.A))
	size += gobuf.VarintSize(int64(s.B))
	return size
}

func (s *AddReq) Marshal(b []byte) int {
	var n int
	n += binary.PutVarint(b[n:], int64(s.A))
	n += binary.PutVarint(b[n:], int64(s.B))
	return n
}

func (s *AddReq) Unmarshal(b []byte) int {
	var n int
	{
		v, x := binary.Varint(b[n:])
		s.A = int(v)
		n += x
	}
	{
		v, x := binary.Varint(b[n:])
		s.B = int(v)
		n += x
	}
	return n
}

func (s *AddReq) MessageID() byte {
	return byte(MsgID_Add)
}

func (s *AddReq) ServiceID() byte {
	return byte(services.ServiceID_Module1)
}

func (s *AddReq) Identity() string {
	return "Module1.AddReq"
}

var _ gobuf.Struct = (*AddRsp)(nil)

func (s *AddRsp) Size() int {
	var size int
	size += gobuf.VarintSize(int64(s.C))
	return size
}

func (s *AddRsp) Marshal(b []byte) int {
	var n int
	n += binary.PutVarint(b[n:], int64(s.C))
	return n
}

func (s *AddRsp) Unmarshal(b []byte) int {
	var n int
	{
		v, x := binary.Varint(b[n:])
		s.C = int(v)
		n += x
	}
	return n
}

func (s *AddRsp) MessageID() byte {
	return byte(MsgID_Add)
}

func (s *AddRsp) ServiceID() byte {
	return byte(services.ServiceID_Module1)
}

func (s *AddRsp) Identity() string {
	return "Module1.AddRsp"
}
