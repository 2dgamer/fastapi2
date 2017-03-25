package module1

import (
	"fastapi2"

	"github.com/funny/link"
)

func (_ *Service) NewRequest(id byte) fastapi2.IMessage {
	switch MessageID(id) {
	case MsgID_Add:
		return &AddReq{}
	}
	return nil
}

func (_ *Service) NewResponse(id byte) fastapi2.IMessage {
	switch MessageID(id) {
	case MsgID_Add:
		return &AddRsp{}
	}
	return nil
}

func (s *Service) HandleRequest(session *link.Session, req fastapi2.IMessage) {
	switch MessageID(req.MessageID()) {
	case MsgID_Add:
		session.Send(s.Add(session, req.(*AddReq)))
	default:
		panic("Unhandled Message Type")
	}
}
