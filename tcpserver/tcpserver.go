package tcpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"os"
	"time"

	"node/database"
	"node/websocket"
)

var Port string = ":9000"
var Nodes map[string]net.Conn = make(map[string]net.Conn)
var Listener net.Listener

type Message struct {
	Command  string
	NodeName string
}

func Run() {
	var err error
	Listener, err = net.Listen("tcp", Port)
	if err != nil {
		fmt.Printf("TCP Listen 실패 : %s\n", err)
		os.Exit(1)
	}

	go HandleListen()
	JoinP2P()
}

func HandleListen() {
	for {
		conn, err := Listener.Accept()
		if err != nil {
			fmt.Printf("신규 노드 연결 승인 실패 : %s\n", err)
			continue
		}

		//신규 노드에게 Message 받기
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("신규 노드 Message 수신 실패 : %s\n", err)
			continue
		}
		var message Message
		err = json.Unmarshal(buffer[:bytesRead], &message)
		if err != nil {
			fmt.Printf("신규 노드 Message 역직렬화 실패 %s\n", err)
			continue
		}

		//"join"이면 다른 노드들에게 전달
		if message.Command == "join" {
			ms := Message{
				Command:  "dial",
				NodeName: message.NodeName,
			}

			msJSON, err := json.Marshal(ms)
			if err != nil {
				fmt.Printf("msJSON 역직렬화 실패 : %s\n", err)
				continue
			}

			for _, conn := range Nodes {

				_, err = conn.Write(msJSON)
				if err != nil {
					fmt.Printf("msJSON 전송 실패 : %s\n", err)
					continue
				}
			}
		}

		fmt.Printf("노드(%s)와 연결\n", message.NodeName)
		Nodes[message.NodeName] = conn

		go HandleNode(message.NodeName)

		files, err := database.GetFiles()
		if err != nil {
			fmt.Printf("에러에렁ㄹ\n")
			continue
		}
		websocket.SendMessages(files, len(Nodes))
	}
}

func HandleNode(remoteName string) {
	defer func() {
		Nodes[remoteName].Close()
		delete(Nodes, remoteName)
		files, err := database.GetFiles()
		if err != nil {
			fmt.Printf("ㅎㄷ%s\n", err)
			return
		}
		websocket.SendMessages(files, len(Nodes))

	}()

	conn := Nodes[remoteName]

	for {
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("노드(%s) 연결 끊김 : %s\n", remoteName, err)
			return
		}

		var message Message
		err = json.Unmarshal(buffer[:bytesRead], &message)
		if err != nil {
			fmt.Printf("노드(%s) Message 역직렬화 실패 %s\n", remoteName, err)
			return
		}

		if message.Command == "dial" {
			newConn, err := net.Dial("tcp", message.NodeName+Port)
			if err != nil {
				fmt.Printf("연결 실패:%s\n", err)
				continue
			}

			ms := Message{
				Command:  "accept",
				NodeName: os.Getenv("HOSTNAME"),
			}

			msJSON, err := json.Marshal(ms)
			if err != nil {
				fmt.Printf("ms 직렬화 실패 :%s\n", err)
				continue
			}

			_, err = newConn.Write(msJSON)
			if err != nil {
				fmt.Printf("msJSON 전송 실패 :%s\n", err)
				continue
			}

			fmt.Printf("노드(%s)와 연결\n", message.NodeName)
			Nodes[message.NodeName] = newConn
			go HandleNode(message.NodeName)

			files, err := database.GetFiles()
			if err != nil {
				fmt.Printf("파팦파파\n")
				return
			}
			fmt.Println("웹소켓 샌드")
			websocket.SendMessages(files, len(Nodes))

		} else if message.Command == "upload" {
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

			fmt.Println("파일 정보 : ", fileInfo)

			// 파일 수신
			uploadDir := "./uploads/"
			os.MkdirAll(uploadDir, os.ModePerm)
			filePath := uploadDir + fileInfo.Name

			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				return
			}

			total := 0
			for {

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

				if fileInfo.Size == total {
					break
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
}

func JoinP2P() {
	targetNodeAddr := os.Getenv("TARGET_NODE")

	if targetNodeAddr == "" {
		return
	}

	conn, err := net.Dial("tcp", targetNodeAddr+Port)
	if err != nil {
		fmt.Printf("타겟 노드에 연결 실패:%s\n", err)
		return
	}

	ms := Message{
		Command:  "join",
		NodeName: os.Getenv("HOSTNAME"),
	}

	msJSON, err := json.Marshal(ms)
	if err != nil {
		fmt.Printf("ms 직렬화 실패:%s\n", err)
		return
	}

	_, err = conn.Write(msJSON)
	if err != nil {
		fmt.Printf("타겟 노드에 msJSON 전송 실패:%s\n", err)
		return
	}

	fmt.Printf("노드(%s)와 연결\n", targetNodeAddr)
	Nodes[targetNodeAddr] = conn
	go HandleNode(targetNodeAddr)
}

func SendFileToOtherNodes(file multipart.File, fileInfo database.File) {
	ms := Message{
		Command:  "upload",
		NodeName: os.Getenv("HOSTNAME"),
	}

	msJSON, err := json.Marshal(ms)
	if err != nil {
		fmt.Printf("ms 직렬화 실패:%s\n", err)
		return
	}

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

			conn.Write(msJSON)

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
