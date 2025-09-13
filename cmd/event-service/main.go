package main

import (
	"github.com/fyzanshaik/bookmyevent-ily/services/event"
)

func main() {
	appConfig, db := event.InitEventService()
	defer db.Close()
	
	event.StartServer(appConfig)
}