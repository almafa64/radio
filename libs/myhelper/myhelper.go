package myhelper

import (
	"radio_site/libs/myconst"
	"radio_site/libs/myfile"
	"radio_site/libs/mystruct"
)

func InvertStatusByte(statuses []byte, pin int) {
    if statuses[pin] == '1' {
        statuses[pin] = '0'
    } else if statuses[pin] == '0' {
        statuses[pin] = '1'
    }
}

func Toggle_pin_status(pin int) []byte {
    statuses := myfile.Read_pin_statuses()

    InvertStatusByte(statuses, pin)

    myfile.Write_pin_file(statuses)
    return statuses
}

func Get_data() []mystruct.Button {
    buttons := make([]mystruct.Button, myconst.MAX_NUMBER_OF_PINS)
    names := myfile.Read_pin_names()
    modes := myfile.Read_pin_modes()

    for i := 0; i < myconst.MAX_NUMBER_OF_PINS; i++ {
        name := names[i]

        buttons[i] = mystruct.Button {
            Name: name,
            Num: i,
            IsToogle: modes[i] == 'T',
        }
    }

    return buttons
}
