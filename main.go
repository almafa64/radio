package main

import (
	"fmt"
	"radio_site/libs/mycamera"
	"radio_site/libs/myconfig"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/myfile"
	"radio_site/libs/myparallel"

	"radio_site/libs/mytpl"
	"radio_site/libs/mywebsocket"

	"log"
	"net/http"
)

func pageHandler(res http.ResponseWriter, req *http.Request) {
    path := req.URL.Path

    if path == "/" {
        index(res)
        return
    }

    http.NotFound(res, req)
}

func index(res http.ResponseWriter) {
    err := mytpl.Tpl.ExecuteTemplate(res, "index.html", nil)

    myerr.CheckErr(err)
}

func main() {
    err := myconfig.LoadOrSaveDefault();
    if err != nil {
        log.Fatalln(err)
    }

    if myconst.MAX_NUMBER_OF_PINS > 63 || myconst.MAX_NUMBER_OF_PINS < 1 {
        log.Fatalln("MAX_NUMBER_OF_PINS cannot be bigger than 63, nor smaller than 1")
    }

    if err := myparallel.CheckPerm(); err == myparallel.ErrPortAccess {
        log.Fatalln(err)
    }

    // if file doesnt exists, create it with default value
    myfile.CheckFile()

    mytpl.TemplateInit()

    if myconfig.Get().Features.Camera {
        mycamera.InitCamera()
    }

    mywebsocket.StartWorker()

    http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
    http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))

    http.HandleFunc("/", pageHandler)
    http.HandleFunc("/radio_ws", mywebsocket.WsHandler)

    webPort := myconfig.Get().WebPort
    log.Printf("Starting HTTP server on :%d", webPort)

    http.ListenAndServe(fmt.Sprintf(":%d", webPort), nil)
}
