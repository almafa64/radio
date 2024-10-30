package main

// #cgo LDFLAGS: -lm
// #include "c_main.h"
import "C"

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
    conn *websocket.Conn
    send chan []byte
}

type Pin struct {
    Num       int
    Status    string
    IsEnabled bool
}

var Tpl *template.Template

const PORT = "8080"
const PIN_FILE_PATH = "pin_status.txt"
const MAX_NUMBER_OF_PINS = 7

const READ_TIMEOUT = 100 * time.Second
const HEARTBEAT_TIMEOUT = 20 * time.Second

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

var (
    clients = make(map[*Client]struct{}) // "hashset"
    clientsLock  sync.Mutex
)

func Template_init() {
    var err error

    Tpl, err = Tpl.ParseGlob("./templates/*.html")

    check_err(err)

    log.Println("Parsed templates:")
    for _, tmpl := range Tpl.Templates() {
        log.Println(" - ", tmpl.Name())
    }
}

func page_handler(res http.ResponseWriter, req *http.Request) {
    path := req.URL.Path

    if path == "/" {
        index(res)
        return
    }

    http.NotFound(res, req)
}

func ws_handler(res http.ResponseWriter, req *http.Request) {
    conn, err := upgrader.Upgrade(res, req, nil)
    if err != nil {
        log.Println("upgrade error:", err)
        return
    }
    
    client := &Client{conn: conn, send: make(chan []byte)}
	defer client.conn.Close()

    addClient(client)
	defer removeClient(client)

	go readMessages(client)

    // wait maximum readTimeout second for pong
    conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT))
    conn.SetPongHandler(func(appData string) error {
        conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT))
        return nil
    })

    // ping every heartbeatTimeout second
    heartbeatTicker := time.NewTicker(HEARTBEAT_TIMEOUT)
    defer heartbeatTicker.Stop()

    for {
		select {
		case <-heartbeatTicker.C:
			// send ping
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(client.conn.RemoteAddr().String(), "ping failed, closing connection:", err)
				return
			}
		case message, ok := <-client.send:
			if !ok {
				// Channel closed, terminate connection
				log.Printf("%s channel closed\n", client.conn.RemoteAddr().String())
				client.conn.WriteMessage(websocket.TextMessage, []byte("closed"))
                return
			}

			// Send the message to the client
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(client.conn.RemoteAddr().String(), "write error, closing connection:", err)
				return
			}
		}
	}
}

func readMessages(client *Client) {
    defer close(client.send)

    clientsLock.Lock()
    client.send <- read_pin_file()
    clientsLock.Unlock()

	for {
		_, message, err := client.conn.ReadMessage()
        if websocket.IsCloseError(err, websocket.CloseGoingAway) {
            return
        } else if err != nil {
			log.Println("Read error:", err)
			return
		}

		clientsLock.Lock()

        // check if message is number and in range of max pin number
        num, err := strconv.Atoi(string(message))
        if err != nil || num >= MAX_NUMBER_OF_PINS {
            continue
        }

        /*
        C.enable_perm()
        C.disable_perm()
        */

		// Send the message to all connected clients
        statuses := toggle_pin_status(num)
		for c := range clients {
            c.send <- statuses
		}

		clientsLock.Unlock()
	}
}

func addClient(client *Client) {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	clients[client] = struct{}{}
	log.Printf("%s connected. Total clients: %d", client.conn.RemoteAddr().String(), len(clients))
}

func removeClient(client *Client) {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	delete(clients, client)
	log.Printf("%s disconnected. Total clients: %d", client.conn.RemoteAddr().String(), len(clients))
}

func index(res http.ResponseWriter) {
    data := gen_pins()
    err := Tpl.ExecuteTemplate(res, "index.html", data)

    check_err(err)
}

func gen_pins() []Pin {
    all_pins := []Pin{}
    for i := 0; i < MAX_NUMBER_OF_PINS; i++ {
        status := get_pin_status(i)

        pin := Pin{
            i + 1,
            status,
            false,
        }

        if status != "" {
            pin.IsEnabled = true
        }

        all_pins = append(all_pins, pin)
    }

    return all_pins
}

func overall_bin_status() int {
    dec_data := 0
    statuses := read_pin_file()
    
    for i := 0; i < MAX_NUMBER_OF_PINS; i++ {
		if statuses[i] == '1' {
			dec_data |= 1 << (MAX_NUMBER_OF_PINS - 1 - i) // set pin bit in dec_data from backward
		}
	}

    return dec_data
}

func get_pin_status(pin int) string {
    status := read_pin_file()[pin]

    switch(status) {
        case '1': return "on"
        case '0': return "off"
        default:  return ""
    }
}

func toggle_pin_status(pin int) []byte {
    statuses := read_pin_file()

    pin_byte := statuses[pin]
    if pin_byte == '1' {
        statuses[pin] = '0'
    } else if pin_byte == '0' {
        statuses[pin] = '1'
    }

    write_pin_file(append(statuses, '\n'))
    return statuses
}

func write_file(filename string, data []byte) {
    err := os.WriteFile(filename, data, 0644)
    check_err(err)
}

func read_file(filename string) []byte {
    data, err := os.ReadFile(filename)
    check_err(err)
    return data[:len(data)-1] // remove newline
}

func write_pin_file(data []byte) {
    write_file(PIN_FILE_PATH, data)
}

func read_pin_file() []byte{
    return read_file(PIN_FILE_PATH)
}

func check_err(e error) {
    if e != nil {
        log.Fatal(e)
    }
}

func main() {
    if MAX_NUMBER_OF_PINS > 63 || MAX_NUMBER_OF_PINS < 1 {
        log.Fatalln("MAX_NUMBER_OF_PINS cant be bigger than 63, nor smaller than 1")
    }

    http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
    http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))

    http.HandleFunc("/", page_handler)
    http.HandleFunc("/radio_ws", ws_handler)

    Template_init()

    // if file doesnt exists, create it with default value
    status_bytes, err := os.ReadFile(PIN_FILE_PATH)
    if os.IsNotExist(err) {
        default_pins := strings.Repeat("-", MAX_NUMBER_OF_PINS) + "\n"
        write_pin_file([]byte(default_pins))
    } else {
        check_err(err)

        // if last byte isnt \n then add it
        if status_bytes[len(status_bytes)-1] != '\n' {
            status_bytes = append(status_bytes, '\n')
            write_pin_file(status_bytes)
        }

        // if pin file length is different than MAX_NUMBER_OF_PINS then modify it
        status_bytes = status_bytes[:len(status_bytes)-1]
        if len(status_bytes) != MAX_NUMBER_OF_PINS {
            statuses := string(status_bytes)
            
            var default_pins string
            if len(statuses) > MAX_NUMBER_OF_PINS {
                default_pins = statuses[:MAX_NUMBER_OF_PINS] + "\n"
            } else {
                default_pins = statuses + strings.Repeat("-", MAX_NUMBER_OF_PINS - len(statuses)) + "\n"
            }

            write_pin_file([]byte(default_pins))
        }
    }

    http.ListenAndServe(":"+PORT, nil)
}
