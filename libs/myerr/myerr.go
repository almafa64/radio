package myerr

import (
	"log"
)

func CheckErrMsg(msg string, e error) {
	if e != nil {
		log.Fatal(msg, e)
	}
}

func Check_err(e error) {
	CheckErrMsg("", e)
}