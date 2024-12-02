package myfile

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"

	"os"
	"strings"
)

func write_file(filename string, pin_statuses string) {
    var data string
    pin_names := Read_pin_names()

    for i := 0; i < len(pin_names); i++ {
        
 
        data += pin_names[i]+ ";" + string(pin_statuses[i]) + "\n"
    }

    err := os.WriteFile(filename, []byte(data), 0644)
    myerr.Check_err(err)
}

func read_file(filepath string) []byte {
    data, err := os.ReadFile(filepath)
    myerr.Check_err(err)
    return data[:len(data)-1] // remove newline
}

func Read_file_lines(filepath string) [][]string {
    var lines [][]string

    data, err := os.ReadFile(filepath)
    myerr.Check_err(err)

    string_data := string(data)

    for _, line := range strings.Split(string_data, "\n") {
        split_line := strings.Split(line, ";")

        lines = append(lines, split_line)
    }

    return lines[:len(lines)-1]
}

func Read_pin_names() []string {
    var pin_names []string
    lines := Read_file_lines(myconst.PIN_FILE_PATH)

    for _, line := range lines {
        pin_names = append(pin_names, strings.TrimSpace(line[0]))
    }

    return pin_names
}

func Read_pin_statuses() []byte {
    pin_statuses := ""
    lines := Read_file_lines(myconst.PIN_FILE_PATH)

    for _, line := range lines {
        pin_statuses += strings.TrimSpace(line[1])
    }

    return []byte(pin_statuses)
}

func Write_pin_file(data []byte) {
    write_file(myconst.PIN_FILE_PATH, string(data))
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
