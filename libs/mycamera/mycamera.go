package mycamera

import (
	"context"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/mywebsocket"

	"github.com/vladimirvivien/go4vl/device"
)

func SendFrames(frames <- chan []byte) {
	for frame := range frames {
		mywebsocket.ClientsLock.Lock()
		for client := range mywebsocket.Clients {
			client.FrameQueue <- frame
		}
		mywebsocket.ClientsLock.Unlock()
	}
}

func InitCamera() *device.Device {
	// Open camera for reading
	camera, err := device.Open(
		myconst.CAMERA_PATH,
		device.WithPixFormat(myconst.CAMERA_FORMAT),
	)

	myerr.CheckErrMsg("failed to open device:", err)

	err = camera.Start(context.Background())
	myerr.CheckErrMsg("camera start:", err)

	go SendFrames(camera.GetOutput())

	return camera
}