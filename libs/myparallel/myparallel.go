package myparallel

// #cgo LDFLAGS: -lm
// #include "../../c_main.h"
import "C"
import (
	"radio_site/libs/myconst"
)

func WritePort(pin_statuses []byte) {
    if !myconst.USE_PARALLEL { return }

    statuses := C.int(0)

    for i, e := range pin_statuses {
        if e != '1' { continue }

        statuses |= 1 << i
    }

    C.set_pins(statuses)
}