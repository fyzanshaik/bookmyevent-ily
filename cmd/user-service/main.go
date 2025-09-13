package main

import (
	"github.com/fyzanshaik/bookmyevent-ily/services/user"
)

func main() {
	appConfig, db := user.InitUserService()
	defer db.Close()
	
	user.StartServer(appConfig)
}
