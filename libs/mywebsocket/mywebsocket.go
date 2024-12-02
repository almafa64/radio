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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	Clients = make(map[*mystruct.Client]struct{}) // "hashset"
	ClientsLock sync.Mutex
)

func startFrameSender(client *mystruct.Client) {
	defer close(client.FrameQueue)
	for frame := range client.FrameQueue {
		if err := client.WriteToClient(websocket.BinaryMessage, frame); err != nil {
			return
		}
	}
}

func addClient(client *mystruct.Client) {
	ClientsLock.Lock()
	defer ClientsLock.Unlock()
	Clients[client] = struct{}{}
	go startFrameSender(client)
	log.Printf("%s connected. Total clients: %d", client.Conn.RemoteAddr().String(), len(Clients))
}

func removeClient(client *mystruct.Client) {
	ClientsLock.Lock()
	defer ClientsLock.Unlock()
	delete(Clients, client)
	log.Printf("%s disconnected. Total clients: %d", client.Conn.RemoteAddr().String(), len(Clients))
}

func Ws_handler(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer conn.Close()

	client := &mystruct.Client{
		Conn: conn,
		Send: make(chan []byte),
		FrameQueue: make(chan []byte, 5),
	}

	addClient(client)
	defer removeClient(client)

	go readMessages(client)

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
			if err := client.WriteToClient(websocket.PingMessage, nil); err != nil {
				log.Println(client.Conn.RemoteAddr().String(), "ping failed, closing connection:", err)
				return
			}
		case message, ok := <-client.Send:
			if !ok {
				// Channel closed, terminate connection
				log.Printf("%s channel closed\n", client.Conn.RemoteAddr().String())
				client.WriteToClient(websocket.TextMessage, []byte("closed"))
				return
			}

			// Send the message to the client
			if err := client.WriteToClient(websocket.TextMessage, message); err != nil {
				log.Println(client.Conn.RemoteAddr().String(), "write error, closing connection:", err)
				return
			}
		}
	}
}

func readMessages(client *mystruct.Client) {
	defer close(client.Send)

	ClientsLock.Lock()
	client.Send <- myfile.Read_pin_statuses()
	ClientsLock.Unlock()

	for {
		msgType, message, err := client.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return
		} else if err != nil {
			log.Println("Read error:", err)
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		// check if message is number and in range of max pin number
		num, err := strconv.Atoi(string(message))
		if err != nil || num >= myconst.MAX_NUMBER_OF_PINS {
			continue
		}

		// Send the message to all connected clients
		statuses := myhelper.Toggle_pin_status(num + 1)
		ClientsLock.Lock()
		for c := range Clients {
			c.Send <- statuses
		}
		ClientsLock.Unlock()
	}
}
