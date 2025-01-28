package myfile

import (
	"errors"
	"io"
	"log"
	"radio_site/libs/myconst"
	"radio_site/libs/myerr"
	"radio_site/libs/myparallel"
	"strconv"
	"sync"

	"os"
	"strings"
)

var ErrRead = errors.New("read error")

var (
    pinFile *os.File
    pinFileLock sync.Mutex
)

var pushButtons = make(map[int]struct{})

func Write_pin_file(pin_statuses []byte) error {
    var data strings.Builder

    pin_names := Read_pin_names()
    if pin_names == nil { return ErrRead }
    pin_modes := Read_pin_modes()

    for i := range len(pin_names) {
        data.WriteString(pin_names[i])
        data.WriteByte(';')
        data.WriteByte(pin_statuses[i])
        data.WriteByte(';')
        data.WriteByte(pin_modes[i])
        data.WriteByte('\n')
    }

    return WriteWholePinFileFD([]byte(data.String()))
}

func Read_file_lines() [][]string {
    data := ReadWholePinFileFD()
    if data == nil { return nil }

    string_data := string(data)
    splitted_lines := strings.Split(string_data, "\n")

    lines := make([][]string, len(splitted_lines))

    for i, line := range splitted_lines{
        lines[i] = strings.Split(line, ";")
    }

    if len(lines)-1 != myconst.MAX_NUMBER_OF_PINS {
        return nil
    }

    return lines[:len(lines)-1] // remove newline
}

func Read_pin_names() []string {
    lines := Read_file_lines()
    if lines == nil { return nil }

    pin_names := make([]string, len(lines))

    for i, line := range lines {
        pin_names[i] = line[0]
    }

    return pin_names
}

func Read_pin_statuses() []byte {
    lines := Read_file_lines()
    if lines == nil { return nil }

    pin_statuses := make([]byte, len(lines))

    for i, line := range lines {
        pin_statuses[i] = line[1][0]
    }

    return pin_statuses
}

func Read_pin_modes() []byte {
    pin_modes := make([]byte, myconst.MAX_NUMBER_OF_PINS)

    for i := range myconst.MAX_NUMBER_OF_PINS {
        if _, o := pushButtons[i]; o {
            pin_modes[i] = 'P'
        } else {
            pin_modes[i] = 'T'
        }
    }

    return pin_modes
}

// ToDo use FD (FileDescription) functions instead
// more performant and safer than reopening the file
// (or store data in memory instead of the file)

func ReadPinFileFD(line int) []byte {
    pinFileLock.Lock()
    defer pinFileLock.Unlock()
    if line == -1 {
        info, err := pinFile.Stat()
        if err != nil {
            log.Println(err)
            return nil
        }
        data := make([]byte, info.Size())

        pinFile.Seek(0, 0)
        _, err = pinFile.Read(data)
        if err != io.EOF && err != nil {
            log.Println(err)
            return nil
        }
        return data
    }

    log.Fatalf("Not implemented!")
    return nil
}

func ReadWholePinFileFD() []byte {
    return ReadPinFileFD(-1)
}

func WritePinFileFD(data []byte, line int) error {
    pinFileLock.Lock()
    defer pinFileLock.Unlock()
    if line == -1 {
        pinFile.Seek(0, 0)
        if _, err := pinFile.Write(data); err != nil {
            log.Println(err)
            return err
        }
        pinFile.Sync()
        return nil
    }

    log.Fatalf("Not implemented!")
    return nil
}

func WriteWholePinFileFD(data []byte) error {
    return WritePinFileFD(data, -1)
}

func print_line_error(msg string, line_num int, line []string) {
    log.Println(msg + " in line #" + strconv.Itoa(line_num) + " '" + strings.Join(line, ";") + "'")
}

func Check_file() {
    first_run := false

    text, err := os.ReadFile(myconst.PIN_FILE_PATH)
    if os.IsNotExist(err) {
        pinFile, err = os.OpenFile(myconst.PIN_FILE_PATH, os.O_RDWR | os.O_CREATE, 0644)
        myerr.Check_err(err)
        text = []byte("button 1;-;T")
        first_run = true
    } else {
        myerr.Check_err(err)
        pinFile, _ = os.OpenFile(myconst.PIN_FILE_PATH, os.O_RDWR, 0644)
    }

    // if last byte isnt \n then add it
    if len(text) == 0  || text[len(text)-1] != '\n' {
        text = append(text, '\n')
        WriteWholePinFileFD(text)
    }

    lines := Read_file_lines()
    for i, line := range lines {
        switch len(line) {
            case 0:
                line = []string{"button " + strconv.Itoa(i + 1), "-", "T"}
            case 1:
                line = []string{line[0], "-", "T"}
            case 2:
                line = []string{line[0], line[1], "T"}
            case 3:
                if line[2] != "T" && line[2] != "P" {
                    print_line_error("Undefined character '" + line[2] + "'", i, lines[i])
                    line[2] = "T"
                }
            default:
                print_line_error("\nToo much part", i, lines[i])
                os.Exit(0)
                log.Fatal()
        }

        if line[1] != "-" && line[1] != "0" && line[1] != "1" {
            print_line_error("Undefined character '" + line[1] + "'", i, lines[i])
            line[1] = "-"
        }

        for j := range line {
            line[j] = strings.TrimSpace(line[j])
        }

        lines[i] = line

        if line[2] == "P" {
            pushButtons[i] = struct{}{}
        }
    }

    linesLen := len(lines)
    if linesLen > myconst.MAX_NUMBER_OF_PINS {
        // remove not needed lines
        log.Println("Removing", (linesLen - myconst.MAX_NUMBER_OF_PINS), "pin lines")
        lines = lines[:myconst.MAX_NUMBER_OF_PINS]
    } else if linesLen < myconst.MAX_NUMBER_OF_PINS {
        // add "button i" lines to fill needed lines
        log.Println("Adding", (myconst.MAX_NUMBER_OF_PINS - linesLen), "pin lines")
        for i := linesLen; i < myconst.MAX_NUMBER_OF_PINS; i++ {
            lines = append(lines, []string{"button " + strconv.Itoa(i + 1), "-", "T"})
        }
    }

    statuses := make([]byte, len(lines))

    // no need to use strings.Builder, only runs at start
    out := ""
    for i, line := range lines {
        out += line[0] + ";" + line[1] + ";" + line[2] + "\n"

        if line[1] == "1" {
            statuses[i] = '1'
        } else {
            statuses[i] = '0'
        }
    }
    myparallel.WritePort(statuses)
    WriteWholePinFileFD([]byte(out))

    if first_run {
        log.Println("Created pins.txt for first time, quiting for configuration.")
        os.Exit(0)
    }
}
