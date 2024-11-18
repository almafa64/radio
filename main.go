package main

// #cgo LDFLAGS: -lm
// #include "c_main.h"
import "C"

import (
	"radio_site/libs/mycamera"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"

	// "radio_site/libs/myfile"
	"radio_site/libs/myhelper"
	"radio_site/libs/mytpl"
	"radio_site/libs/mywebsocket"

	"log"
	"net/http"
)

func page_handler(res http.ResponseWriter, req *http.Request) {
    path := req.URL.Path

    if path == "/" {
        index(res)
        return
    }

    http.NotFound(res, req)
}

func index(res http.ResponseWriter) {
    data := myhelper.Get_data()
    err := mytpl.Tpl.ExecuteTemplate(res, "index.html", data)

    myerr.Check_err(err)
}
    
func main() {
    if myconst.MAX_NUMBER_OF_PINS > 63 || myconst.MAX_NUMBER_OF_PINS < 1 {
        log.Fatalln("MAX_NUMBER_OF_PINS cant be bigger than 63, nor smaller than 1")
    }

    // if file doesnt exists, create it with default value
    //myfile.Check_file()
    
    mytpl.Template_init()

    camera := mycamera.InitCamera()
    defer camera.Close()

    http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
    http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))

    http.HandleFunc("/", page_handler)
    http.HandleFunc("/radio_ws", mywebsocket.Ws_handler)
    http.HandleFunc("/video", mycamera.Streaming)
    
    http.ListenAndServe(":"+myconst.PORT, nil)
}
