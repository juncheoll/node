package broad

import (
	"fmt"
	"net"
	"os"
)

var port string = "12345"

var conn *net.UDPConn
var listenConn *net.UDPConn

func Readybroadcast() {
	var err error

	listenAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Println("브로드캐스트 리스너 Resolve 실패", err)
		return
	}

	listenConn, err = net.ListenUDP("udp", listenAddr)
	if err != nil {
		fmt.Println("브로드캐스트 리스너 생성 실패", err)
		return
	}

}

func Send() {
	broadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+port)
	if err != nil {
		fmt.Println("브로드캐스트 Resolve 실패", err)
		return
	}

	conn, err = net.DialUDP("udp", nil, broadcastAddr)
	if err != nil {
		fmt.Println("UDP 소켓 생성 실패", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(os.Getenv("HOSTNAME")))
	if err != nil {
		fmt.Println("브로드캐스트 전송 실패:", err)
		return
	}
}

func Receive() string {
	buffer := make([]byte, 1024)
	n, _, err := listenConn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("UDP 소켓 읽기 실패:", err)
		return ""
	}
	recvMessage := string(buffer[:n])
	fmt.Println("UDP 수신 :", recvMessage)

	return recvMessage
}
