package myerr

import (
    "log"
)

func Check_err(e error) {
    if e != nil {
        log.Fatal(e)
    }
}
