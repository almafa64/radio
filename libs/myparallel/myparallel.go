package myparallel

// #include <sys/io.h>
import "C"
import (
	"radio_site/libs/appstate"
	"radio_site/libs/myconfig"

	"errors"
	"log"
)

const LPT_IO_PORT = C.ushort(0x378)

var ErrParallelNotEnabled = errors.New("Parallel not enabled")
var ErrPortAccess = errors.New("Access denied to parallel port")

func WritePort(pin_states appstate.PinStates) {
	if err := CheckPerm(); err != nil {
		if err == ErrPortAccess {
			log.Println(err)
		}
		return
	}

	status_bits := C.uchar(0)

	for pin, enabled := range pin_states {
		if !enabled {
			continue
		}

		status_bits |= 1 << pin
	}

	C.outb(status_bits, LPT_IO_PORT)
}

func CheckPerm() error {
	if !myconfig.Get().Features.Parallel {
		return ErrParallelNotEnabled
	}

	if C.ioperm(C.ulong(LPT_IO_PORT), 1, 1) != 0 {
		log.Printf("[LPT] Port 0x%X is not accessible (missing root privileges?)\n", LPT_IO_PORT)
		return ErrPortAccess
	}

	return nil
}
