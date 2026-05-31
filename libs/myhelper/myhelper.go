package myhelper

func InvertBit(data uint64, pin uint64) uint64 {
	return data ^ (1 << pin)
}

func SetBitTo(data uint64, pin uint64, state bool) uint64 {
	var num uint64 = 0
	if state {
		num = 1
	}

	return (data & ^(1 << pin)) | (num << pin)
}

func InvertStatusByte(statuses []byte, pin int) {
	switch statuses[pin] {
	case '1':
		statuses[pin] = '0'
	case '0':
		statuses[pin] = '1'
	}
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
