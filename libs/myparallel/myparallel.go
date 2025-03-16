package myparallel

// #include <sys/io.h>
import "C"
import (
	"errors"
	"log"
	"radio_site/libs/myconfig"
)

const LPT_IO_PORT = C.ushort(0x378)

var ErrParallelNotEnabled = errors.New("Parallel not enabled")
var ErrPortAccess = errors.New("Access denied to parallel port")

func WritePort(pin_statuses []byte) {
    if err := CheckPerm(); err != nil {
        if err == ErrPortAccess { log.Println(err) }
        return
    }

    status_bits := C.uchar(0)

    for i, e := range pin_statuses {
        if e != '1' { continue }

        status_bits |= 1 << i
    }

    C.outb(status_bits, LPT_IO_PORT)
}

func CheckPerm() (error) {
    if !myconfig.Get().Peripheral.Parallel { return ErrParallelNotEnabled }
    if C.ioperm(C.ulong(LPT_IO_PORT), 1, 1) != 0 {
        log.Printf("[LPT] Port 0x%X is not accessible (missing root privileges?)\n", LPT_IO_PORT)
        return ErrPortAccess
    }
    return nil
}