package mycamera

import (
	"radio_site/libs/myconfig"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/mystruct"
	"radio_site/libs/mywebsocket"

	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

var stopChannel chan struct{} = nil

func shouldStop() bool {
	select {
	case _, ok := <-stopChannel:
		// this will only run when channel closes, because there shouldnt be any sending in this channel
		return !ok
	default:
		return false
	}
}

func openCamera(camera **device.Device, config myconfig.CameraModule) bool {
	var format uint32
	switch config.Format {
	case "mjpeg":
		format = v4l2.PixelFmtMJPEG
	default:
		log.Fatalf("[camera/%s] Unsupported image format: %s", config.Name, config.Format)
	}

	resolution := strings.Split(config.Resolution, "x")
	width := myerr.CheckTup(strconv.ParseInt(resolution[0], 10, 32))
	height := myerr.CheckTup(strconv.ParseInt(resolution[1], 10, 32))

	pix_fmt := v4l2.PixFormat{
		PixelFormat: format,
		Width:       uint32(width),
		Height:      uint32(height),
	}

	retry := false
	var err error

	for {
		if *camera != nil {
			(*camera).Close()
			*camera = nil
		}

		if shouldStop() {
			return false
		}

		*camera, err = device.Open(
			config.Device,
			device.WithPixFormat(pix_fmt),
			device.WithFPS(config.Fps),
		)

		if err == nil {
			log.Printf("[camera/%s] %s opened", config.Name, config.Device)

			return true
		} else if !retry {
			log.Printf("[camera/%s] device open: %s", config.Name, err)

			retry = true
			time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
		}
	}
}

func startCamera(camera *device.Device, config myconfig.CameraModule) bool {
	retry := false

	for {
		camera.Stop()

		if shouldStop() {
			return false
		}

		if err := camera.Start(context.Background()); err == nil {
			log.Printf("[camera/%s] %s connected", config.Name, config.Device)

			return true
		} else if !retry {
			log.Printf("[camera/%s] device start: %s", config.Name, err)

			retry = true
			time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
		}
	}
}

func frameLoop(camera *device.Device, id int) bool {
	timeoutClock := time.NewTicker(myconst.CAMERA_FRAME_TIMEOUT)

	for {
		if shouldStop() {
			return false
		}

		select {
		case frame := <-camera.GetFrames():
			prepmsg, _ := websocket.NewPreparedMessage(websocket.BinaryMessage, append([]byte{byte(id)}, frame.Data...))
			frame.Release()

			mywebsocket.Clients.Range(func(key, value any) bool {
				key.(*mystruct.Client).PrepMessageQueue <- prepmsg
				return true
			})

			timeoutClock.Reset(myconst.CAMERA_FRAME_TIMEOUT)
		case <-timeoutClock.C:
			return true
		}
	}
}

func cameraWorker(id int, config myconfig.CameraModule) {
	var camera *device.Device

	for {
		if !openCamera(&camera, config) {
			break
		}

		if !startCamera(camera, config) {
			break
		}

		if !frameLoop(camera, id) {
			break
		}

		log.Printf("[camera/%s] %s disconnected", config.Name, config.Device)
		camera.Close()
	}

	if camera != nil {
		log.Printf("[camera/%s] %s exited", config.Name, config.Device)
		camera.Close()
	}
}

func InitCamera() {
	cameraCounter := 0

	if stopChannel != nil {
		close(stopChannel)
		time.Sleep(myconst.CAMERA_FRAME_TIMEOUT * 2) // leave time for goroutines to exit
	}

	stopChannel = make(chan struct{})

	log.Printf("[camera] initialising cameras")

	for _, segment := range myconfig.Get().Segments {
		for _, module := range segment {
			camera, ok := module.(myconfig.CameraModule)
			if !ok {
				continue
			}

			go cameraWorker(cameraCounter, camera)
			cameraCounter++
		}
	}
}
