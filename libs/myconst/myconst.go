package myconst

import (
	"time"
)

const READ_TIMEOUT = 100 * time.Second
const HEARTBEAT_TIMEOUT = 20 * time.Second
const CAMERA_TRY_TIMEOUT = 2 * time.Second
const CAMERA_FRAME_TIMEOUT = 1500 * time.Millisecond
