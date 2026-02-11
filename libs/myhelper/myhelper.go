package myhelper

import (
	"radio_site/libs/myconfig"
	"radio_site/libs/myfile"
	"radio_site/libs/mystruct"
)

func InvertStatusByte(statuses []byte, pin int) {
	switch statuses[pin] {
	case '1':
		statuses[pin] = '0'
	case '0':
		statuses[pin] = '1'
	}
}

func TogglePinStatus(pin int) []byte {
	statuses := myfile.ReadPinStatuses()
	if statuses == nil {
		return nil
	}

	InvertStatusByte(statuses, pin)

	if err := myfile.WritePinFile(statuses); err != nil {
		return nil
	}
	return statuses
}

func GetData() []mystruct.Button {
	buttons := make([]mystruct.Button, myconfig.GetButtonCount())

	names := myfile.ReadPinNames()
	if names == nil {
		return nil
	}
	modes := myfile.ReadPinModes()

	for i := 0; i < myconfig.GetButtonCount(); i++ {
		name := names[i]

		buttons[i] = mystruct.Button{
			Name:     name,
			Num:      i,
			IsToggle: modes[i] == 'T',
		}
	}

	return buttons
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
