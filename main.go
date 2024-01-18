package main

import (
	"fmt"
	"os"
	"strconv"

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

	args := os.Args
	if len(args) < 2 {
		fmt.Println("포트 번호를 입력하세요.")
		os.Exit(1)
	}

	port := args[1]
	_, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println("포트 번호를 정수로 입력하세요.")
		os.Exit(1)
	}

	httpserver.Port = port
	tcpserver.Port, _ = tcpserver.ConvertHTTPToTCPPort(port)

	database.SetDB(port)
	tcpserver.Run(port)
	httpserver.Run()
}
