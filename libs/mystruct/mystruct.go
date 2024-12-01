package mystruct

import (
    "github.com/gorilla/websocket"
)

type Client struct {
    Conn      *websocket.Conn
    Send      chan []byte
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
