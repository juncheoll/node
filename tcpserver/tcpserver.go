package tcpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"

	"node/database"
	"node/websocket"
)

var Port string = ":9000"
var Nodes map[string]net.Conn
var Listener net.Listener

func Run() {

	Nodes = make(map[string]net.Conn)
	var err error
	Listener, err = net.Listen("tcp", Port)
	if err != nil {
		fmt.Printf("TCP Listen 실패 : %s\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			conn, err := Listener.Accept()
			if err != nil {
				fmt.Printf("신규 노드 연결 승인 실패 : %s\n", err)
				continue
			}

			//신규 노드 이름 받아오기
			buffer := make([]byte, 512)
			bytesRead, err := conn.Read(buffer)
			if err != nil {
				fmt.Printf("신규 노드 포트 수신 실패 : %s\n", err)
				continue
			}
			newNodeName := string(buffer[:bytesRead])

			fmt.Printf("노드(%s)와 연결\n", newNodeName)
			Nodes[newNodeName] = conn

			go HandleNode(newNodeName)
		}
	}()

	JoinP2P()
}

func HandleNode(remoteName string) {
	defer func() {
		Nodes[remoteName].Close()
		delete(Nodes, remoteName)
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
		websocket.SendMessages(files)
	}
}

func JoinP2P() {
	values := url.Values{}
	values.Add("dname", os.Getenv("HOSTNAME"))

	url := "http://nodecenter:8080/?" + values.Encode()

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("노드 리스트 받아오기 실패:%s\n", err)
		return
	}
	defer response.Body.Close()

	var nodeList []string
	err = json.NewDecoder(response.Body).Decode(&nodeList)
	if err != nil {
		fmt.Printf("노드 리스트 디코딩 실패:%s\n", err)
		return
	}

	//nodeList에 각각 dial
	for _, nodeName := range nodeList {
		conn, err := net.Dial("tcp", nodeName+Port)
		if err != nil {
			fmt.Printf("노드(%s)에 연결 실패:%s\n", nodeName, err)
			continue
		}

		fmt.Printf("노드(%s)와 연결\n", nodeName)
		Nodes[nodeName] = conn
		go HandleNode(nodeName)

		//연결된 노드에 본인 포트 전달
		conn.Write([]byte(os.Getenv("HOSTNAME")))
	}

	fmt.Printf("P2P 네트워크 입장 성공\n")
}

func SendFileToOtherNodes(file multipart.File, fileInfo database.File) {
	for nodeName, conn := range Nodes {

		fmt.Printf("%s 에게 %s전송\n", nodeName, fileInfo.Name)

		//파일 정보 전송
		fmt.Println("파일 정보 전달 : ", fileInfo)

		jsonData, err := json.Marshal(fileInfo)
		if err != nil {
			fmt.Println("파일 정보 직렬화 실패:", err)
			return
		}
		conn.Write([]byte(jsonData))

		//파일 내용 전송
		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			continue
		}

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
	}
	fmt.Println("전체 노드 전송 완료")
}
