package myhelper

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myfile"
	"radio_site/libs/mystruct"
)

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
    buttons := make([]mystruct.Button, myconst.MAX_NUMBER_OF_PINS)
    data := myfile.Read_pin_names()

    for i := 0; i < myconst.MAX_NUMBER_OF_PINS; i++ {
        name := data[i]

        buttons[i] = mystruct.Button {
            Name: name,
            Num: i,
        }
    }

    return buttons
}
