package main

import (
	"github.com/fyzanshaik/bookmyevent-ily/services/booking"
)

func main() {
	appConfig, db := booking.InitBookingService()
	defer db.Close()

	booking.StartServer(appConfig)
}
