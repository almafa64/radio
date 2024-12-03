package myhelper

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myfile"
	"radio_site/libs/mystruct"
)

func Overall_bin_status() int {
    dec_data := 0
    statuses := myfile.Read_pin_statuses()
    
    for i := 0; i < myconst.MAX_NUMBER_OF_PINS; i++ {
		if statuses[i] == '1' {
			dec_data |= 1 << (myconst.MAX_NUMBER_OF_PINS - 1 - i) // set pin bit in dec_data from backward
		}
	}

    return dec_data
}

func Get_pin_status(pin int) string {
    status := myfile.Read_pin_statuses()[pin]

    switch(status) {
        case '1': return "on"
        case '0': return "off"
        default:  return ""
    }
}

func Toggle_pin_status(pin int) []byte {
    statuses := myfile.Read_pin_statuses()

    pin_byte := statuses[pin]
    if pin_byte == '1' {
        statuses[pin] = '0'
    } else if pin_byte == '0' {
        statuses[pin] = '1'
    }

    myfile.Write_pin_file(statuses)
    return statuses
}

func Get_data() []mystruct.Button {
    buttons := []mystruct.Button{}
    data := myfile.Read_pin_names()

    for i := 0; i < myconst.MAX_NUMBER_OF_PINS; i++ {
        name := data[i]

        button := mystruct.Button {
            Name: name,
            Num: i,
        }

        buttons = append(buttons, button)
    }

    return buttons
}
