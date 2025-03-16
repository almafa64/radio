package mycamera

import (
	"context"
	"log"
	"radio_site/libs/myconfig"
	"radio_site/libs/myconst"
	"radio_site/libs/mystruct"
	"radio_site/libs/mywebsocket"
	"strconv"
	"strings"
	"time"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func sendFrames(frames <-chan []byte) {
	for frame := range frames {
		mywebsocket.Clients.Range(func(key, value any) bool {
			key.(*mystruct.Client).FrameQueue <- frame
			return true
		})
	}
}

// try connecting camera every 5 seconds
func tryConnectCamera(config myconfig.Camera) {
	had_err := false

	resolution := strings.Split(config.Resolution, "x");
	var format uint32
	switch config.Format {
	case "mjpeg":
		format = v4l2.PixelFmtMJPEG

	default:
		log.Fatalf("[camera] Unsupported image format: %s", config.Format)
	};

	width, err := strconv.ParseInt(resolution[0], 10, 32)
	if err != nil {
		log.Fatalln(err)
	}
	height, err := strconv.ParseInt(resolution[1], 10, 32)
	if err != nil {
		log.Fatalln(err)
	}

	pix_fmt := v4l2.PixFormat{
		PixelFormat: format,
		Width: uint32(width),
		Height: uint32(height),
	}

	var camera *device.Device
	for {
		camera, err = device.Open(
			config.Device,
			device.WithPixFormat(pix_fmt),
			device.WithFPS(config.Fps),
		)

		if err == nil {
			break
		} else if !had_err { // only print first error to reduce clutter
			log.Println("open device:", err)
			had_err = true
		}

		time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
	}

	log.Println("Successfuly opened", camera.Name())

	had_err = false

	for {
		if err = camera.Start(context.Background()); err == nil {
			go sendFrames(camera.GetOutput())
			return
		} else if !had_err { // only print first error to reduce clutter
			log.Println("camera start:", err)
			had_err = true
		}

		time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
	}
}

func InitCamera() {
	if len(myconfig.Get().Camera) == 0 {
		return
	}
	go tryConnectCamera(myconfig.Get().Camera[0])
}
