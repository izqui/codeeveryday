package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/izqui/helpers"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
)

var pin gpio.Pin

var contributions int

func init() {

	contributions = 0
}
func main() {

	pin = setupGPIO()

	nchan := make(chan int)

	user = flag.String("user", "izqui", "Github Username for contributions graph")
	flag.Parse()

	go monitor(user, nchan)
	go listen(nchan)

	//Program will be forever waiting for this channel to be sent data
	noexit := make(chan int)
	<-noexit
	panic("Suicide")
}

func setupGPIO() gpio.Pin {

	p, err := gpio.OpenPin(rpi.GPIO25, gpio.ModeOutput)
	if err != nil {
		panic(err)
	}

	// turn the led off on exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Printf("\nClearing and unexporting the pin.\n")
			pin.Clear()
			pin.Close()
			os.Exit(0)
		}
	}()

	return p
}

func monitor(username string, nchan chan int) {

	time.Sleep(5 * time.Second)

	for {

		res, err := http.Get("https://github.com/users/" + username + "/contributions_calendar_data")
		if err == nil {

			var data []interface{}

			helpers.DecodeJSON(res.Body, &data)

			n := int(data[len(data)-1].([]interface{})[1].(float64)) //Send last value, index 1
			fmt.Println(n)

			nchan <- n

		} else {

			fmt.Println(err)
		}

		time.Sleep(time.Second)
	}
}

func listen(nchan chan int) {

	for {

		select {

		case n := <-nchan:
			if n != contributions {

				//Number of contributions has changed
				contributions = n

				pin.Set()
			}
		}
		if contributions == 0 {

			pin.Set()
			time.Sleep(time.Second / 2)
			pin.Clear()
			time.Sleep(time.Second / (3 / 4))
		}
	}
}
