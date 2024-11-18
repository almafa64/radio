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
		device.WithPixFormat(myconst.CAMERA_FORMAT),
	)

	myerr.CheckErrMsg("failed to open device:", err)

	err = camera.Start(context.TODO())
	myerr.CheckErrMsg("camera start:", err)

	frames = camera.GetOutput()
	return camera
}