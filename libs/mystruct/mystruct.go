package mystruct

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
    Conn *websocket.Conn
    Send chan []byte
    ConnLock sync.Mutex
    FrameQueue chan []byte
}

func (client *Client) WriteToClient(messageType int, data []byte) error {
    client.ConnLock.Lock()
    defer client.ConnLock.Unlock()
    return client.Conn.WriteMessage(messageType, data)
}

type Pin struct {
    Num       int
    Status    string
    IsEnabled bool
}

type IndexTemplate struct {
    Pins []Pin
    UseCamera bool
}