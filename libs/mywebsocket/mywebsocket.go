package mywebsocket

// #include "../../c_main.h"
import "C"
import (
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/myfile"
	"radio_site/libs/myhelper"
	"radio_site/libs/myparallel"
	"radio_site/libs/mystruct"
	"strings"

	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		host := r.Header.Get("X-Host")
		if host == "" {
			host = r.Host
		}

		if strings.Contains(host, "localhost") {
			return true
		}

		u, err := url.Parse(r.Header["Origin"][0])
		if err != nil {
			myerr.CheckErrMsg("origin error: ", err)
			return false
		}

		return strings.EqualFold(u.Hostname(), host)
	},
}

var (
	Clients = make(map[*mystruct.Client]struct{}) // "hashset"
	ClientsLock sync.Mutex
)

var ButtonsHeld sync.Map

func clientsToString() string {
	var builder strings.Builder
	for client := range Clients {
		builder.WriteString(client.Name)
		builder.WriteByte(',')
	}
	return builder.String()
}

func holdingClientsToString() string {
	var builder strings.Builder
	ButtonsHeld.Range(func(key, value any) bool {
		builder.WriteString(value.(*mystruct.Client).Name)
		builder.WriteByte(';')
		builder.WriteString(strconv.Itoa(key.(int)))
		builder.WriteByte(',')
		return true
	})
	return builder.String()
}

func applyHeldButtons(statuses []byte) []byte {
	ButtonsHeld.Range(func(key, value any) bool {
		pin := key.(int)
		myhelper.InvertStatusByte(statuses, pin)
		return true
	})
	return statuses
}

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
	delete(Clients, client)
	users := clientsToString()
	ClientsLock.Unlock()

	ButtonsHeld.Range(func(key, value any) bool {
		if value != client { return true }

		ButtonsHeld.Delete(key)
		return false
	})

	usersHolding := holdingClientsToString()

	log.Printf("%s disconnected. Total clients: %d", client.Conn.RemoteAddr().String(), len(Clients))
	broadcast([]byte("h" + usersHolding))
	broadcast([]byte("u" + users))
	broadcast(applyHeldButtons(myfile.Read_pin_statuses()))
}

func broadcast(text []byte) {
	ClientsLock.Lock()
	for c := range Clients {
		c.Send <- text
	}
	ClientsLock.Unlock()
}

func Ws_handler(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer conn.Close()

	name := req.Header.Get("X-User")
	if name == "" {
		log.Println(req.RemoteAddr, " has no name")
		name = req.RemoteAddr
	} else {
		name += "(" + req.RemoteAddr + ")"
	}

	client := &mystruct.Client{
		Conn: conn,
		Send: make(chan []byte),
		FrameQueue: make(chan []byte, 5),
		Name: name,
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

	if myconst.USE_PARALLEL && !C.enable_perm() {
        log.Println("Failed to get access to port for '" + client.Name + "'")
		return
    }

	ClientsLock.Lock()
	client.Send <- applyHeldButtons(myfile.Read_pin_statuses())
	users := clientsToString()
	ClientsLock.Unlock()

	usersHolding := holdingClientsToString()
	broadcast([]byte("h" + usersHolding))
	broadcast([]byte("u" + users))

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
		pin, err := strconv.Atoi(string(message))
		if err != nil || pin >= myconst.MAX_NUMBER_OF_PINS {
			continue
		}

		modes := myfile.Read_pin_modes()
		isToggleButton := modes[pin] == 'T'

		var statuses []byte

		if !isToggleButton {
			statuses = myfile.Read_pin_statuses()
			value, loaded := ButtonsHeld.LoadOrStore(pin, client)
			if value != client { continue }
			if loaded {
				ButtonsHeld.Delete(pin)
			}
			usersHolding := holdingClientsToString()
			broadcast([]byte("h" + usersHolding))
		} else {
			statuses = myhelper.Toggle_pin_status(pin)
		}

		applyHeldButtons(statuses)
		myparallel.WritePort(statuses)
		broadcast(statuses)
	}
}
