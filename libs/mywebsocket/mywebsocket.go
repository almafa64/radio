package mywebsocket

import (
	"radio_site/libs/appstate"
	"radio_site/libs/myconfig"
	"radio_site/libs/myconst"
	"radio_site/libs/myparallel"
	"radio_site/libs/mystruct"

	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	holdingCommandPrefix  = "h"
	userListCommandPrefix = "u"
	editorCommandPrefix   = "e"
	jsonCommandPrefix     = "j"
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
			log.Println("origin error: ", err)
			return false
		}

		return strings.EqualFold(u.Hostname(), host)
	},
	EnableCompression: true,
}

var (
	Clients      sync.Map
	ClientCount  atomic.Int64 // sync.Map doesnt have any len function
	editorClient *mystruct.Client
)

var IncomingCameraFrames chan mystruct.CameraFrame = make(chan mystruct.CameraFrame, 20)

type wrap[T any] struct {
	Event string
	Data  T
}

func clientsToString() string {
	var builder strings.Builder

	Clients.Range(func(key, value any) bool {
		builder.WriteString(key.(*mystruct.Client).Name)
		builder.WriteByte(',')
		return true
	})

	return builder.String()
}

func holdingClientsToString() string {
	var builder strings.Builder

	Clients.Range(func(key, value any) bool {
		client := key.(*mystruct.Client)
		
		if client.HoldingPinNumber == -1 {
			return true
		}

		builder.WriteString(client.Name)
		builder.WriteByte(';')
		builder.WriteString(strconv.Itoa(client.HoldingPinNumber))
		builder.WriteByte(',')
		
		return true
	})

	return builder.String()
}

func frameSender(client *mystruct.Client) {
	for prepMessage := range client.PrepMessageQueue {
		client.ConnLock.Lock()
		client.Conn.WritePreparedMessage(prepMessage)
		client.ConnLock.Unlock()
	}
}

func createPinEvent(pin_states appstate.PinStates) []byte {
	return pin_states.ToByteSlice()
}

func createUserEvent() []byte {
	users := clientsToString()
	return []byte(userListCommandPrefix + users)
}

func createJSONEvent[T any](data T, eventName string) []byte {
	out, err := json.Marshal(wrap[T]{eventName, data})
	
	if err != nil {
		log.Printf("%v", err)
		return nil
	}

	return append([]byte(jsonCommandPrefix), out...)
}

func createHolderEvent() []byte {
	users := holdingClientsToString()
	return []byte(holdingCommandPrefix + users)
}

func createEditorEvent(editor_name string) []byte {
	return []byte(editorCommandPrefix + editor_name)
}

func broadcast(data []byte) {
	Clients.Range(func(key, value any) bool {
		key.(*mystruct.Client).Send <- data
		return true
	})
}

func setEditor(client *mystruct.Client) {
	if client == nil {
		editorClient = nil
		broadcast(createEditorEvent(""))
		return
	}

	editorClient = client
	broadcast(createEditorEvent(client.Name))
}

func addClient(client *mystruct.Client) {
	Clients.Store(client, struct{}{})
	ClientCount.Add(1)

	client.Conn.EnableWriteCompression(true)

	go readMessages(client)
	go frameSender(client)

	log.Printf("%s connected. Total clients: %d", client.Name, ClientCount.Load())
}

func removeClient(client *mystruct.Client) {
	Clients.Delete(client)
	ClientCount.Add(-1)

	client.Conn.Close()
	close(client.PrepMessageQueue)

	broadcast(createUserEvent())

	if client.HoldingPinNumber != -1 {
		pin := client.HoldingPinNumber

		client.HoldingPinNumber = -1
		broadcast(createHolderEvent())

		state := appstate.GetWritable()
		state.PinStates.TogglePinStatus(pin)
		broadcast(createPinEvent(state.PinStates))
		state.Release()
	}

	log.Printf("%s disconnected. Total clients: %d", client.Name, ClientCount.Load())

	if editorClient == client {
		setEditor(nil)
	}
}

func WsHandler(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	name := req.Header.Get("X-User")
	if name == "" {
		name = req.Header.Get("X-Real-IP")
		if name == "" {
			name = req.RemoteAddr
		}
		log.Println(name, "has no name")
	} else {
		name += " (" + req.Header.Get("X-Real-IP") + ")"
	}

	client := &mystruct.Client{
		Conn:             conn,
		Send:             make(chan []byte),
		Name:             name,
		PrepMessageQueue: make(chan *websocket.PreparedMessage, 5),
		HoldingPinNumber: -1,
	}

	addClient(client)
	defer removeClient(client)

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
				log.Println(client.Name, "ping failed. Closing connection:", err)
				return
			}
		case message, ok := <-client.Send:
			if !ok {
				// Channel closed, terminate connection
				log.Println(client.Name, "channel closed")
				return
			}

			// Send the message to the client
			if err := client.WriteToClient(websocket.TextMessage, message); err != nil {
				log.Println(client.Name, "write error, closing connection:", err)
				return
			}
		}
	}
}

// TODO: pushing 2 push button fast enough after each other stucks one
// TODO: 2 client pushing same push buttons stucks it

func readMessages(client *mystruct.Client) {
	defer close(client.Send)

	client.Send <- createJSONEvent(myconfig.Get().Segments, "page_scheme")

	client.Send <- []byte(userListCommandPrefix + "*" + client.Name)

	state := appstate.Get()
	statuses := state.PinStates.ToByteSlice()
	state.Release()

	client.Send <- statuses
	client.Send <- createHolderEvent()

	broadcast(createUserEvent())

	if editorClient != nil {
		client.Send <- createEditorEvent(editorClient.Name)
	}

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

		if message[0] == editorCommandPrefix[0] {
			switch editorClient {
			case nil:
				setEditor(client)
			case client:
				setEditor(nil)
			}

			continue
		}

		if message[0] == jsonCommandPrefix[0] {
			var a wrap[json.RawMessage]
			if err := json.Unmarshal(message[1:], &a); err != nil {
				log.Println(client.Name, "error in json:", err)
				continue
			}
			if a.Event == "" {
				log.Println(client.Name, "strange json:", string(message[1:]))
				continue
			}

			switch a.Event {
			case "page_scheme":
				var segments myconfig.Segments
				if err = json.Unmarshal(a.Data, &segments); err != nil {
					log.Println(client.Name, "error in json:", err, string(a.Data))
					break
				}
				// TODO: save page and broadcast
			}
			continue
		}

		// check if message is number
		pin, err := strconv.Atoi(string(message))
		if err != nil {
			continue
		}

		button := myconfig.Get().GetButtonByPin(pin)
		if button == nil {
			continue
		}

		isToggleButton := button.IsToggle

		if !isToggleButton {
			// if button is not held by requesting user, deny it
			if client.HoldingPinNumber != -1 && pin != client.HoldingPinNumber {
				continue
			}

			if client.HoldingPinNumber != -1 { // if button already held by requesting user, release it
				client.HoldingPinNumber = -1
			} else {
				client.HoldingPinNumber = pin
			}

			broadcast(createHolderEvent())
		}

		state := appstate.GetWritable()

		state.PinStates.TogglePinStatus(pin)

		myparallel.WritePort(state.PinStates)
		broadcast(createPinEvent(state.PinStates))

		state.Release()
	}
}
