package main

import (
	"fmt"
	"time"
)

type logwriter struct {
}

func (writer logwriter) Write(bytes []byte) (int, error) {
	return fmt.Println(time.Now().Format("15:04:05" + " " + string(bytes)))
}
