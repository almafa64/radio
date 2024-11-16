package myfile

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"

	"os"
	"strings"
)

func write_file(filename string, data []byte) {
    err := os.WriteFile(filename, data, 0644)
    myerr.Check_err(err)
}

func read_file(filepath string) []byte {
    data, err := os.ReadFile(filepath)
    myerr.Check_err(err)
    return data[:len(data)-1] // remove newline
}

func Write_pin_file(data []byte) {
   write_file(myconst.PIN_FILE_PATH, data)
}

func Read_pin_file() []byte{
    return read_file(myconst.PIN_FILE_PATH)
}

func Check_file() {
    status_bytes, err := os.ReadFile(myconst.PIN_FILE_PATH)
    if os.IsNotExist(err) {
        default_pins := strings.Repeat("-", myconst.MAX_NUMBER_OF_PINS) + "\n"
        Write_pin_file([]byte(default_pins))
    } else {
        myerr.Check_err(err)

        // if last byte isnt \n then add it
        if status_bytes[len(status_bytes)-1] != '\n' {
            status_bytes = append(status_bytes, '\n')
            Write_pin_file(status_bytes)
        }

        // if pin file length is different than MAX_NUMBER_OF_PINS then modify it
        status_bytes = status_bytes[:len(status_bytes)-1]
        if len(status_bytes) != myconst.MAX_NUMBER_OF_PINS {
            statuses := string(status_bytes)
            
            var default_pins string
            if len(statuses) > myconst.MAX_NUMBER_OF_PINS {
                default_pins = statuses[:myconst.MAX_NUMBER_OF_PINS] + "\n"
            } else {
                default_pins = statuses + strings.Repeat("-", myconst.MAX_NUMBER_OF_PINS - len(statuses)) + "\n"
            }

            Write_pin_file([]byte(default_pins))
        }
    }
}
