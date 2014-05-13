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

var redPin, greenPin gpio.Pin

func main() {

	redPin, greenPin = setupGPIO()

	user := flag.String("user", "izqui", "Github Username for contributions graph")
	flag.Parse()
	fmt.Printf("User: %s\n", *user)

	//MAIN
	nchan := make(chan int)
	go monitor(*user, nchan)
	go listen(nchan)

	//Program will be forever waiting for this channel to be sent data
	noexit := make(chan int)
	<-noexit
	panic("Suicide")
}

func setupGPIO() (gpio.Pin, gpio.Pin) {

	p1, err := gpio.OpenPin(rpi.GPIO17, gpio.ModeOutput)
	if err != nil {
		panic(err)
	}

	p2, err := gpio.OpenPin(rpi.GPIO21, gpio.ModeOutput)
	if err != nil {
		panic(err)
	}

	// turn the led off on exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Printf("\nClearing and unexporting the pin.\n")
			p1.Clear()
			p1.Close()
			p2.Clear()
			p2.Close()
			os.Exit(0)
		}
	}()

	return p1, p2
}

func monitor(username string, nchan chan int) {

	for {

		res, err := http.Get("https://github.com/users/" + username + "/contributions_calendar_data")
		if err == nil {

			var data []interface{}

			helpers.DecodeJSON(res.Body, &data)

			n := int(data[len(data)-1].([]interface{})[1].(float64)) //Send last value, index 1

			nchan <- n

		} else {

			fmt.Println(err)
		}

		time.Sleep(time.Second)
	}
}

func listen(nchan chan int) {

	contributions := int(0)
	light(redPin, greenPin)
	for {

		select {

		case n := <-nchan:
			if n != contributions {

				//Number of contributions has changed
				contributions = n

				fmt.Printf("Contributions today: %i \n", contributions)

				if contributions > 0 {

					light(greenPin, redPin)
				} else {
					light(redPin, greenPin)
				}
			}
		}
	}
}

func light(on, off gpio.Pin) {

	off.Clear()
	on.Set()

}
