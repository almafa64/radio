package mycamera

import (
	"context"
	"log"
	"radio_site/libs/myconfig"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/mystruct"
	"radio_site/libs/mywebsocket"
	"strconv"
	"strings"
	"time"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func cameraWorker(id int, config myconfig.CameraModule) {
	var format uint32
	switch config.Format {
	case "mjpeg":
		format = v4l2.PixelFmtMJPEG
	default:
		log.Fatalf("[camera/%s] Unsupported image format: %s", config.Name, config.Format)
	};

	resolution := strings.Split(config.Resolution, "x");
	width := myerr.CheckTup(strconv.ParseInt(resolution[0], 10, 32))
	height := myerr.CheckTup(strconv.ParseInt(resolution[1], 10, 32))

	pix_fmt := v4l2.PixFormat{
		PixelFormat: format,
		Width: uint32(width),
		Height: uint32(height),
	}

	var err error
	var camera *device.Device

	for {
		retry := false
		for {
			camera, err = device.Open(
				config.Device,
				device.WithPixFormat(pix_fmt),
				device.WithFPS(config.Fps),
			)
			if err == nil {
				retry = false
				log.Printf("Opened %s", config.Name)
				break
			} else if !retry {
				log.Printf("[camera/%s] device open: %s", config.Name, err)
				retry = true
			}
			time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
		}

		for {
			if err := camera.Start(context.Background()); err == nil {
				retry = false
				log.Printf("Started %s", config.Name)
				break
			} else if !retry {
				log.Printf("[camera/%s] device start: %s", config.Name, err)
				retry = true
			}
			time.Sleep(myconst.CAMERA_TRY_TIMEOUT)
		}

		timeoutClock := time.NewTicker(myconst.CAMERA_FRAME_TIMEOUT)

		log.Printf("[camera/%s] %s connected", config.Name, config.Device)
		FRAMELOOP: for {
			select {
			case frame := <- camera.GetOutput():
				mywebsocket.Clients.Range(func(key, value any) bool {
					mywebsocket.IncomingCameraFrames <- mystruct.CameraFrame{
						CamId: uint8(id),
						Data:  frame,
					}
					return true
				})
				timeoutClock.Reset(myconst.CAMERA_FRAME_TIMEOUT)
			case <-timeoutClock.C:
				break FRAMELOOP
			}
		}
		log.Printf("[camera/%s] %s disconnected", config.Name, config.Device)
		camera.Stop()
	}
}

func InitCamera() {
	cameraCounter := 0

	for _, segment := range myconfig.Get().Segments {
		for _, module := range segment {
			if module.GetType() != "cam" { continue }
			go cameraWorker(cameraCounter, module.(myconfig.CameraModule))
			cameraCounter++
		}
	}
}
