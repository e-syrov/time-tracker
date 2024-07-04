package main

import (
	"fmt"
	"time-tracker/internal/app"
)

func main() {
	err := app.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
}
