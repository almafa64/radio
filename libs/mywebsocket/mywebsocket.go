package mywebsocket

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myfile"
	"radio_site/libs/myhelper"
	"radio_site/libs/mystruct"

	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	Clients = make(map[*mystruct.Client]struct{}) // "hashset"
	ClientsLock sync.Mutex
)

func AddClient(client *mystruct.Client) {
	ClientsLock.Lock()
	defer ClientsLock.Unlock()
	Clients[client] = struct{}{}
	log.Printf("%s connected. Total clients: %d", client.Conn.RemoteAddr().String(), len(Clients))
}

func RemoveClient(client *mystruct.Client) {
	ClientsLock.Lock()
	defer ClientsLock.Unlock()
	delete(Clients, client)
	log.Printf("%s disconnected. Total clients: %d", client.Conn.RemoteAddr().String(), len(Clients))
}

func WriteToClient(client *mystruct.Client, messageType int, data []byte) error {
	client.ConnLock.Lock()
	defer client.ConnLock.Unlock()
	return client.Conn.WriteMessage(messageType, data)
}

func Ws_handler(res http.ResponseWriter, req *http.Request) {
	conn, err := Upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	client := &mystruct.Client{Conn: conn, Send: make(chan []byte)}
	defer client.Conn.Close()

	AddClient(client)
	defer RemoveClient(client)

	go ReadMessages(client)

	// wait maximum readTimeout second for pong
	conn.SetReadDeadline(time.Now().Add(myconst.READ_TIMEOUT))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(myconst.READ_TIMEOUT))
		return nil
	})

	// ping every heartbeatTimeout second
	heartbeatTicker := time.NewTicker(myconst.HEARTBEAT_TIMEOUT)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-heartbeatTicker.C:
			// send ping
			if err := WriteToClient(client, websocket.PingMessage, nil); err != nil {
				log.Println(client.Conn.RemoteAddr().String(), "ping failed, closing connection:", err)
				return
			}
		case message, ok := <-client.Send:
			if !ok {
				// Channel closed, terminate connection
				log.Printf("%s channel closed\n", client.Conn.RemoteAddr().String())
				WriteToClient(client, websocket.TextMessage, []byte("closed"))
				return
			}

			// Send the message to the client
			if err := WriteToClient(client, websocket.TextMessage, message); err != nil {
				log.Println(client.Conn.RemoteAddr().String(), "write error, closing connection:", err)
				return
			}
		}
	}
}

func ReadMessages(client *mystruct.Client) {
	defer close(client.Send)

	ClientsLock.Lock()
	client.Send <- myfile.Read_pin_file()
	ClientsLock.Unlock()

	for {
		_, message, err := client.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return
		} else if err != nil {
			log.Println("Read error:", err)
			return
		}

		ClientsLock.Lock()

		// check if message is number and in range of max pin number
		num, err := strconv.Atoi(string(message))
		if err != nil || num >= myconst.MAX_NUMBER_OF_PINS {
			continue
		}

		// Send the message to all connected clients
		statuses := myhelper.Toggle_pin_status(num)
		log.Println("Statuses:", statuses)
		for c := range Clients {
			c.Send <- statuses
		}

		ClientsLock.Unlock()
	}
}
