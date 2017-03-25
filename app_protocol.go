package fastapi2

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/funny/link"
)

func (app *App) newClientCodec(rw io.ReadWriter) (link.Codec, error) {
	return app.newCodec(rw, app.newResponse), nil
}

func (app *App) newServerCodec(rw io.ReadWriter) (link.Codec, error) {
	return app.newCodec(rw, app.newRequest), nil
}

func (app *App) newCodec(rw io.ReadWriter, newMessage func(byte, byte) (IMessage, error)) link.Codec {
	c := &codec{
		app:        app,
		conn:       rw.(net.Conn),
		reader:     bufio.NewReaderSize(rw, app.ReadBufSize),
		newMessage: newMessage,
	}
	c.headBuf = c.headData[:]
	return c
}

func (app *App) newRequest(serviceID, messageID byte) (IMessage, error) {
	if service := app.services[serviceID]; service != nil {
		if msg := service.NewRequest(messageID); msg != nil {
			return msg, nil
		}
		return nil, DecodeError{fmt.Sprintf("Unsupported Message Type: [%d, %d]", serviceID, messageID)}
	}
	return nil, DecodeError{fmt.Sprintf("Unsupported Service: [%d, %d]", serviceID, messageID)}
}

func (app *App) newResponse(serviceID, messageID byte) (IMessage, error) {
	if service := app.services[serviceID]; service != nil {
		if msg := service.NewResponse(messageID); msg != nil {
			return msg, nil
		}
		return nil, DecodeError{fmt.Sprintf("Unsupported Message Type: [%d, %d]", serviceID, messageID)}
	}
	return nil, DecodeError{fmt.Sprintf("Unsupported Service: [%d, %d]", serviceID, messageID)}
}

type EncodeError struct {
	Message interface{}
}

func (encodeError EncodeError) Error() string {
	return fmt.Sprintf("Encode Error: %v", encodeError.Message)
}

type DecodeError struct {
	Message interface{}
}

func (decodeError DecodeError) Error() string {
	return fmt.Sprintf("Decode Error: %v", decodeError.Message)
}

const packetHeadSize = 4 + 2

type codec struct {
	app        *App
	headBuf    []byte
	headData   [packetHeadSize]byte
	conn       net.Conn
	reader     *bufio.Reader
	newMessage func(byte, byte) (IMessage, error)
}

func (c *codec) Conn() net.Conn {
	return c.conn
}

func (c *codec) Receive() (msg interface{}, err error) {
	if c.app.RecvTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.app.RecvTimeout))
		defer c.conn.SetReadDeadline(time.Time{})
	}

	if _, err = io.ReadFull(c.reader, c.headBuf); err != nil {
		return
	}

	packetSize := int(binary.LittleEndian.Uint32(c.headBuf))

	if packetSize > c.app.MaxRecvSize {
		return nil, DecodeError{fmt.Sprintf("Too Large Receive Packet Size: %d", packetSize)}
	}

	packet := make([]byte, packetSize)

	if _, err = io.ReadFull(c.reader, packet); err == nil {
		msg1, err1 := c.newMessage(c.headData[4], c.headData[5])
		if err1 == nil {
			func() {
				defer func() {
					if panicErr := recover(); panicErr != nil {
						err = DecodeError{panicErr}
					}
				}()
				msg1.Unmarshal(packet)
			}()
			msg = msg1
		} else {
			err = err1
		}
	}
	return
}

func (c *codec) Send(m interface{}) (err error) {
	msg := m.(IMessage)

	packetSize := msg.Size()

	if packetSize > c.app.MaxSendSize {
		panic(EncodeError{fmt.Sprintf("Too Large Send Packet Size: %d", packetSize)})
	}

	packet := make([]byte, packetHeadSize+packetSize)
	binary.LittleEndian.PutUint32(packet, uint32(packetSize))
	packet[4] = msg.ServiceID()
	packet[5] = msg.MessageID()

	func() {
		defer func() {
			if panicErr := recover(); panicErr != nil {
				err = EncodeError{panicErr}
			}
		}()
		msg.Marshal(packet[packetHeadSize:])
	}()

	if c.app.SendTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.app.SendTimeout))
		defer c.conn.SetWriteDeadline(time.Time{})
	}

	_, err = c.conn.Write(packet)
	return
}

func (c *codec) Close() error {
	return c.conn.Close()
}
