package myconst

import (
    "time"
)

const PORT = "8080"
const MAX_NUMBER_OF_PINS = 7

const PIN_FILE_PATH = "./pins.txt"

const READ_TIMEOUT = 100 * time.Second
const HEARTBEAT_TIMEOUT = 20 * time.Second
