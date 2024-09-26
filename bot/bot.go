// myrobot/robot.go
package bot

import (
	"fmt"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

func NewRaspiRobot() *gobot.Robot {
	r := raspi.NewAdaptor()
	led := gpio.NewLedDriver(r, "7")

	work := func() {
		gobot.Every(1*time.Second, func() {
			led.Toggle()
			fmt.Println("LED toggled")
		})
	}

	robot := gobot.NewRobot("raspiBot",
		[]gobot.Connection{r},
		[]gobot.Device{led},
		work,
	)

	return robot
}
