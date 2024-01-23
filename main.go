package main

import (
	"fmt"

	"node/database"
	"node/httpserver"
	"node/tcpserver"
)

func main() {
	defer func() {
		fmt.Printf("종료")
		database.DB.Close()
	}()

	database.SetDB()
	tcpserver.Run()
	httpserver.Run()

}
