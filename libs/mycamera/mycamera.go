package mycamera

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

var (
	frames <-chan []byte
)

func Streaming(w http.ResponseWriter, req *http.Request) {
	mimeWriter := multipart.NewWriter(w)
	w.Header().Set("Content-Type", fmt.Sprintf("multipart/x-mixed-replace; boundary=%s", mimeWriter.Boundary()))
	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	var frame []byte
	for frame = range frames {
		partWriter, err := mimeWriter.CreatePart(partHeader)
		if err != nil {
			log.Printf("failed to create multi-part writer: %s", err)
			return
		}

		if _, err := partWriter.Write(frame); err != nil {
			log.Printf("failed to write image: %s", err)
		}
	}
}

func InitCamera() *device.Device {
	camera, err := device.Open(
		myconst.CAMERA_PATH,
		device.WithPixFormat(v4l2.PixFormat{PixelFormat: v4l2.PixelFmtMJPEG, Width: 1280, Height: 720}),
	)

	myerr.CheckErrMsg("failed to open device:", err)

	err = camera.Start(context.TODO())
	myerr.CheckErrMsg("camera start:", err)

	frames = camera.GetOutput()
	return camera
}

/*
func main() {
	port := ":3000"
	devName := "/dev/video0"
	flag.StringVar(&devName, "d", devName, "device name (path)")
	flag.StringVar(&port, "p", port, "webcam service port")

	camera, err := device.Open(
		devName,
		device.WithPixFormat(v4l2.PixFormat{PixelFormat: v4l2.PixelFmtMJPEG, Width: 1280, Height: 720}),
	)
	if err != nil {
		log.Fatalf("failed to open device: %s", err)
	}
	defer camera.Close()

	if err := camera.Start(context.TODO()); err != nil {
		log.Fatalf("camera start: %s", err)
	}

	frames = camera.GetOutput()

	log.Printf("Serving images: [0.0.0.0%s/video]", port)
	
	http.HandleFunc("/video", imageServ)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, "index")
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
*/