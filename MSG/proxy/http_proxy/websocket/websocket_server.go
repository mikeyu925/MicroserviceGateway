package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

func main() {
	var addr = "localhost:8002"
	http.HandleFunc("/wshandler", whHandler)
	log.Println("Starting websocket server at " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func whHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil) // 协议升级
	if err != nil {
		log.Print("upgrade error")
		return
	}
	defer conn.Close()

	go func() {
		// 服务器主动向客户端推送消息
		for {
			err := conn.WriteMessage(websocket.TextMessage, []byte("heart beat"))
			if err != nil {
				return
			}
			time.Sleep(3 * time.Second)
		}
	}()

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			log.Print("read error: ", err)
			break
		}
		fmt.Printf("receive msg : %s\n", msg)
		newMsg := string(msg) + "哈哈哈你好"
		msg = []byte(newMsg)
		err = conn.WriteMessage(mt, msg)
		if err != nil {
			log.Print("write error : ", err)
			break
		}
	}
}
