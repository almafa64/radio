package mystruct

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
    Conn *websocket.Conn
    Send chan []byte
    ConnLock sync.Mutex
}

type Pin struct {
    Num       int
    Status    string
    IsEnabled bool
}

type Button struct {
    Name      string
    Num       int
}

type IndexTemplate struct {
    Pins []Pin
    UseCamera bool
}