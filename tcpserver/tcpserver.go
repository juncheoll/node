package tcpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"os"
	"time"

	"node/broad"
	"node/database"
	"node/websocket"
)

var Port string = ":9000"
var Nodes map[string]net.Conn = make(map[string]net.Conn)

type Message struct {
	Command  string
	NodeName string
}

func Run() {
	ln, err := net.Listen("tcp", Port)
	if err != nil {
		fmt.Printf("TCP Listen 실패 : %s\n", err)
		os.Exit(1)
	}

	go handleListen(ln)
	go listenBroadcast()
}

func handleListen(ln net.Listener) {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("노드 연결 승인 실패 : %s\n", err)
			continue
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("노드 수신 실패 : %s\n", err)
			continue
		}

		remoteName := string(buffer[:n])
		connectNode(remoteName, conn)
	}
}

func listenBroadcast() {
	broad.Send()
	broad.Readybroadcast()

	for {
		nodeName := broad.Receive()

		conn, err := net.Dial("tcp", nodeName+Port)
		if err != nil {
			fmt.Printf("연결 실패:%s\n", err)
			continue
		}

		_, err = conn.Write([]byte(os.Getenv("HOSTNAME")))
		if err != nil {
			fmt.Println("쓰기 실패:", err)
			continue
		}

		connectNode(nodeName, conn)
	}
}

func connectNode(nodeName string, conn net.Conn) {
	fmt.Printf("노드(%s)와 연결\n", nodeName)
	Nodes[nodeName] = conn
	go handleNode(nodeName)

	files, err := database.GetFiles()
	if err != nil {
		fmt.Printf("파팦파파\n")
		return
	}
	fmt.Println("웹소켓 샌드")
	websocket.SendMessages(files, len(Nodes))

}

func handleNode(remoteName string) {
	defer func() {
		Nodes[remoteName].Close()
		delete(Nodes, remoteName)
		fmt.Printf("노드(%s) 연결 끊김\n", remoteName)
		files, err := database.GetFiles()
		if err != nil {
			fmt.Printf("ㅎㄷ%s\n", err)
			return
		}
		websocket.SendMessages(files, len(Nodes))

	}()

	conn := Nodes[remoteName]

	for {
		// 파일 정보 수신
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("노드(%s) 연결 끊김 : %s\n", remoteName, err)
			return
		}

		var fileInfo database.File
		err = json.Unmarshal(buffer[:bytesRead], &fileInfo)
		if err != nil {
			fmt.Println("파일 정보 역직렬화 실패:", err)
			return
		}

		// 파일 수신
		uploadDir := "./uploads/"
		os.MkdirAll(uploadDir, os.ModePerm)
		filePath := uploadDir + fileInfo.Name

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			return
		}

		for total := 0; total != fileInfo.Size; {

			buffer = make([]byte, 1048576)
			n, err := conn.Read(buffer)
			total += n
			if err != nil {
				fmt.Printf("파일 수신 실패 : %s\n", err)
				return
			}

			_, err = file.Write(buffer[:n])
			if err != nil {
				fmt.Printf("파일 쓰기 실패 : %s\n", err)
				return
			}
		}

		file.Close()
		if err != nil {
			fmt.Printf("Error saving file content: %v\n", err)
			return
		}

		_, err = database.SaveFileToDB(fileInfo.Name, filePath)
		if err != nil {
			fmt.Printf("database saveing fail : %v\n", err)
			return
		}

		fmt.Printf("File '%s' saved\n", fileInfo.Name)

		files, err := database.GetFiles()
		if err != nil {
			fmt.Printf("파팦파파\n")
			return
		}
		fmt.Println("웹소켓 샌드")
		websocket.SendMessages(files, len(Nodes))
	}
}

func SendFileToOtherNodes(file multipart.File, fileInfo database.File) {
	jsonData, err := json.Marshal(fileInfo)
	if err != nil {
		fmt.Println("파일 정보 직렬화 실패:", err)
		return
	}

	for nodeName, conn := range Nodes {
		func(nodeName string, conn net.Conn) {
			timer := time.NewTicker(1000 * time.Microsecond)
			defer timer.Stop()

			fmt.Printf("%s 에게 %s전송\n", nodeName, fileInfo.Name)

			//파일 정보 전송
			fmt.Println("파일 정보 전달 : ", fileInfo)

			<-timer.C
			conn.Write(jsonData)

			//파일 내용 전송
			_, err = file.Seek(0, 0)
			if err != nil {
				fmt.Println(err)
				return
			}

			<-timer.C
			buffer := make([]byte, 1048576)
			for {

				n, err := file.Read(buffer)
				if err == io.EOF {
					fmt.Println("파일 끝")
					_, err = conn.Write(buffer[:n])
					if err != nil {
						fmt.Printf("파일 전송 실패 : %s\n", err)
						return
					}
					break
				} else if err != nil {
					fmt.Printf("파일 읽기 실패 : %s\n", err)
					return
				}

				_, err = conn.Write(buffer[:n])
				if err != nil {
					fmt.Printf("파일 전송 실패 : %s\n", err)
					return
				}
			}

			fmt.Printf("File sent to node %s\n", nodeName)
		}(nodeName, conn)
	}
	fmt.Println("전체 노드 전송 완료")
}
