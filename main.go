package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/izqui/helpers"
	"github.com/stianeikeland/go-rpio"
)

var pin rpio.Pin
var contributions int

func init() {

	contributions = 0
}
func main() {

	setupGPIO()

	nchan := make(chan int)

	go monitor(nchan)
	go listen(nchan)

	//Program will be forever waiting for this channel to be sent data
	noexit := make(chan int)
	<-noexit
	panic("Suicide")
}

func setupGPIO() {

	pin := rpio.Pin(25)

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	pin.Output()
}

func monitor(nchan chan int) {

	for {

		res, err := http.Get("https://github.com/users/izqui/contributions_calendar_data")
		if err == nil {

			var data []interface{}

			helpers.DecodeJSON(res.Body, &data)

			n := int(data[len(data)-1].([]interface{})[1].(float64)) //Send last value, index 1
			fmt.Println(n)

			nchan <- n

		} else {

			fmt.Println(err)
		}

		time.Sleep(time.Second / 2)
	}
}

func listen(nchan chan int) {

	for {

		select {

		case n := <-nchan:
			if n != contributions {

				//Number of contributions has changed
				contributions = n

				pin.High()
			}
		}
		fmt.Println("For")
		if contributions == 0 {

			pin.Toggle()
			time.Sleep(time.Second / 2)
		}
	}
}
