package main

import (
	"radio_site/libs/appstate"
	"radio_site/libs/mycamera"
	"radio_site/libs/myconfig"
	"radio_site/libs/myerr"
	"radio_site/libs/myparallel"
	"radio_site/libs/mytpl"
	"radio_site/libs/mywebsocket"

	"fmt"
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
	if err := myconfig.LoadOrSaveDefault(); err != nil {
		log.Fatalln(err)
	}

	if err := appstate.LoadOrSaveDefault(); err != nil {
		log.Fatalln(err)
	}

	if myconfig.Get().GetButtonCount() > 63 || myconfig.Get().GetButtonCount() < 1 {
		log.Fatalln("There cannot be more buttons than 63, nor less than 1")
	}

	if err := myparallel.CheckPerm(); err == myparallel.ErrPortAccess {
		log.Fatalln(err)
	}

	mytpl.TemplateInit()

	if myconfig.Get().Features.Camera {
		mycamera.InitCamera()
	}

	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
	http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))

	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/radio_ws", mywebsocket.WsHandler)

	webPort := myconfig.Get().WebPort
	log.Printf("Starting HTTP server on :%d", webPort)

	http.ListenAndServe(fmt.Sprintf(":%d", webPort), nil)
}
