package myhelper

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myfile"
	"radio_site/libs/mystruct"
)

func Gen_pins() []mystruct.Pin {
    all_pins := []mystruct.Pin{}
    for i := 0; i < myconst.MAX_NUMBER_OF_PINS; i++ {
        status := Get_pin_status(i)

        pin := mystruct.Pin{
            Num: i + 1,
            Status: status,
            IsEnabled: false,
        }

        if status != "" {
            pin.IsEnabled = true
        }

        all_pins = append(all_pins, pin)
    }

    return all_pins
}

func Overall_bin_status() int {
    dec_data := 0
    statuses := myfile.Read_pin_file()
    
    for i := 0; i < myconst.MAX_NUMBER_OF_PINS; i++ {
		if statuses[i] == '1' {
			dec_data |= 1 << (myconst.MAX_NUMBER_OF_PINS - 1 - i) // set pin bit in dec_data from backward
		}
	}

    return dec_data
}

func Get_pin_status(pin int) string {
    status := myfile.Read_pin_file()[pin]

    switch(status) {
        case '1': return "on"
        case '0': return "off"
        default:  return ""
    }
}

func Toggle_pin_status(pin int) []byte {
    statuses := myfile.Read_pin_file()

    pin_byte := statuses[pin]
    if pin_byte == '1' {
        statuses[pin] = '0'
    } else if pin_byte == '0' {
        statuses[pin] = '1'
    }

    myfile.Write_pin_file(append(statuses, '\n'))
    return statuses
}
