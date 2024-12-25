package myparallel

// #cgo LDFLAGS: -lm
// #include "../../c_main.h"
import "C"
import (
	"log"
	"radio_site/libs/myconst"
)

func WritePort(pin_statuses []byte) {
    if !myconst.USE_PARALLEL { return }

    if !C.enable_perm() {
        log.Println("Failed to get access to port!")
        return
    }

    status_bits := C.uchar(0)

    for i, e := range pin_statuses {
        if e != '1' { continue }

        status_bits |= 1 << i
    }

    C.set_pins(status_bits)
}