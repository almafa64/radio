package mystruct

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn             *websocket.Conn
	ConnLock         sync.Mutex
	Send             chan []byte
	Name             string
	PrepMessageQueue chan *websocket.PreparedMessage
	HoldingPinNumber int
}

func (client *Client) WriteToClient(messageType int, data []byte) error {
	client.ConnLock.Lock()
	defer client.ConnLock.Unlock()
	return client.Conn.WriteMessage(messageType, data)
}

type CameraFrame struct {
	CamId uint8
	Data  []byte
}

type Button struct {
	Name     string
	Num      int
	IsToggle bool
}

type IndexTemplate struct {
	Buttons []Button
}
