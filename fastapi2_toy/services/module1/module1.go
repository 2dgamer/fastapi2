package module1

import "github.com/funny/link"

type Module1 struct {
}

type AddReq struct {
	A int
	B int
}

type AddRsp struct {
	C int
}

func (_ *Module1) Add(session *link.Session, req *AddReq) *AddRsp {
	return &AddRsp{
		req.A + req.B,
	}
}
