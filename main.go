package main

import (
	"fmt"

	"node/database"
	"node/httpserver"
	"node/tcpserver"
)

var Name string

func main() {
	defer func() {
		fmt.Printf("종료")
		database.DB.Close()
		tcpserver.Listener.Close()
	}()

	database.SetDB()
	tcpserver.Run()
	httpserver.Run()
}
