package myfile

import (
	"io"
	"log"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"strconv"
	"sync"

	"os"
	"strings"
)

var (
    pinFile *os.File
    pinFileLock sync.Mutex
)

func write_file(filename string, pin_statuses []byte) {
    var data strings.Builder
    pin_names := Read_pin_names()

    for i := 0; i < len(pin_names); i++ {
        data.WriteString(pin_names[i])
        data.WriteByte(';')
        data.WriteByte(pin_statuses[i])
        data.WriteByte('\n')
    }

    err := os.WriteFile(filename, []byte(data.String()), 0644)
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
    lines := Read_file_lines(myconst.PIN_FILE_PATH)
    pin_statuses := make([]byte, len(lines))

    for i, line := range lines {
        pin_statuses[i] = strings.TrimSpace(line[1])[0]
    }

    return pin_statuses
}

func Write_pin_file(data []byte) {
    write_file(myconst.PIN_FILE_PATH, data)
}

func Read_pin_file() []byte{
    return read_file(myconst.PIN_FILE_PATH)
}

// ToDo use FD (FileDescription) functions instead
// more performant and safer than reopening the file
// (or store data in memory instead of the file)

func ReadPinFileFD(line int) []byte {
    pinFileLock.Lock()
    defer pinFileLock.Unlock()
    if line == -1 {
        info, err := pinFile.Stat()
        myerr.Check_err(err)
        data := make([]byte, info.Size())

        pinFile.Seek(0, 0)
        _, err = pinFile.Read(data)
        if err != io.EOF {
            myerr.Check_err(err)
        }
        return data
    }

    log.Fatalf("Not implemented!")
    return nil
}

func ReadWholePinFileFD() []byte {
    return ReadPinFileFD(-1)
}

func WritePinFileFD(data []byte, line int) {
    pinFileLock.Lock()
    defer pinFileLock.Unlock()
    if line == -1 {
        pinFile.Seek(0, 0)
        _, err := pinFile.Write(data)
        myerr.Check_err(err)
        pinFile.Sync()
        return
    }

    log.Fatalf("Not implemented!")
}

func WriteWholePinFileFD(data []byte) {
    WritePinFileFD(data, -1)
}

func Check_file() {
    status_bytes, err := os.ReadFile(myconst.PIN_FILE_PATH)
    if os.IsNotExist(err) {
        pinFile, err = os.OpenFile(myconst.PIN_FILE_PATH, os.O_RDWR | os.O_CREATE, 0644)
        myerr.Check_err(err)
        status_bytes = []byte("button 1;-")
    } else {
        myerr.Check_err(err)
        pinFile, _ = os.OpenFile(myconst.PIN_FILE_PATH, os.O_RDWR, 0644)
    }

    // if last byte isnt \n then add it
    if len(status_bytes) == 0  || status_bytes[len(status_bytes)-1] != '\n' {
        status_bytes = append(status_bytes, '\n')
        WriteWholePinFileFD(status_bytes)
    }

    lines := Read_file_lines(myconst.PIN_FILE_PATH)
    for i, line := range lines {
        if len(line) != 2 {
            lines[i] = []string{"button " + strconv.Itoa(i + 1), "-"}
        }
    }

    linesLen := len(lines)
    if linesLen == myconst.MAX_NUMBER_OF_PINS {
        return
    }

    if linesLen > myconst.MAX_NUMBER_OF_PINS {
        // remove not needed lines
        lines = lines[:myconst.MAX_NUMBER_OF_PINS]
    } else if linesLen < myconst.MAX_NUMBER_OF_PINS {
        // add "button i" lines to fill needed lines
        for i := linesLen; i < myconst.MAX_NUMBER_OF_PINS; i++ {
            lines = append(lines, []string{"button " + strconv.Itoa(i + 1), "-"})
        }
    }

    // no need to use strings.Builder, only runs at start
    out := ""
    for _, line := range lines {
        out += line[0] + ";" + line[1] + "\n"
    }
    WriteWholePinFileFD([]byte(out))
}
