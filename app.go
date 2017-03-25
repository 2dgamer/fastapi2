package fastapi2

import (
	"log"
	"net"
	"runtime/debug"
	"time"

	"github.com/funny/link"
)

type APIs map[byte][2]interface{}

type IService interface {
	ServiceID() byte
	NewRequest(byte) IMessage
	NewResponse(byte) IMessage
	HandleRequest(*link.Session, IMessage)
}

type IMessage interface {
	ServiceID() byte
	MessageID() byte
	Identity() string
	Size() int
	Marshal(b []byte) int
	Unmarshal(b []byte) int
}

type Handler interface {
	InitSession(*link.Session) error
	Transaction(*link.Session, IMessage, func())
}

type App struct {
	services []IService

	ReadBufSize  int
	SendChanSize int
	MaxRecvSize  int           // 最大接收字节
	MaxSendSize  int           // 最大发送字节
	RecvTimeout  time.Duration // 接收超时
	SendTimeout  time.Duration // 发送超时
}

func New() *App {
	return &App{
		services:     make([]IService, 256),
		ReadBufSize:  1024,
		SendChanSize: 1024,
		MaxRecvSize:  64 * 1024,
		MaxSendSize:  64 * 1024,
	}
}

func (app *App) handleSession(session *link.Session, handler Handler) {
	defer session.Close()

	if handler.InitSession(session) != nil {
		return
	}

	for {
		msg, err := session.Receive()
		if err != nil {
			return
		}

		req := msg.(IMessage)
		handler.Transaction(session, req, func() {
			app.services[req.ServiceID()].HandleRequest(session, req)
		})
	}
}

func (app *App) Dial(network, address string) (*link.Session, error) {
	return link.Dial(network, address, link.ProtocolFunc(app.newClientCodec), app.SendChanSize)
}

func (app *App) Listen(network, address string, handler Handler) (*link.Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return app.NewServer(listener, handler), nil
}

func (app *App) NewClient(conn net.Conn) *link.Session {
	codec, _ := app.newClientCodec(conn)
	return link.NewSession(codec, app.SendChanSize)
}

func (app *App) NewServer(listener net.Listener, handler Handler) *link.Server {
	if handler == nil {
		handler = &noHandler{}
	}

	return link.NewServer(listener, link.ProtocolFunc(app.newServerCodec), app.SendChanSize,
		link.HandlerFunc(func(session *link.Session) {
			app.handleSession(session, handler)
		}))
}

func (app *App) Register(service IService) {
	app.services[service.ServiceID()] = service
}

type noHandler struct {
}

func (t *noHandler) InitSession(session *link.Session) error {
	return nil
}

func (t *noHandler) Transaction(session *link.Session, req IMessage, work func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("fastapi2: unhandled panic when processing '%s' - '%s'", req.Identity(), err)
			log.Println(string(debug.Stack()))
		}
	}()
	work()
}
