package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

type (
	ucaActionMessage struct {
		Action string `json:"action"`
	}

	ucaStatusReply struct {
		Action  string  `json:"action"`
		Status  string  `json:"status"`
		Type    *int    `json:"type"`
		Version *string `json:"version"`
	}
)

var (
	ucaPort serial.Port

	ucaUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ucaClients = map[*websocket.Conn]bool{}

	ucaByte2Status = map[byte]string{
		0x11: "idle",
		0x12: "received",
		0x13: "reserved",
		0x14: "disabled",
		0x15: "reserved",
		0x16: "sensor_error",
		0x17: "fishing",
		0x18: "checksum_error",
		0x19: "reserved",
		0x4b: "rejected",
		0x50: "accepted",
	}

	ucaAction2Byte = map[string]byte{
		"enable":  0x01,
		"disable": 0x02,
		"info":    0x03,
		"status":  0x11,
	}
)

func startUCA(device string) {
	var err error
	ucaPort, err = serial.Open(device, &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err == nil {
		log.Println("opened uca device", device)
	} else {
		log.Fatal(err)
	}
	go acceptCoins()
	http.HandleFunc("/ict/uca", ucaServer)
}

func ucaServer(w http.ResponseWriter, r *http.Request) {
	c, err := ucaUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	defer func() {
		delete(ucaClients, c)
		c.Close()
	}()
	ucaClients[c] = true
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Println("read error:", err)
			}
			break
		}
		var msg ucaActionMessage
		if err = json.Unmarshal(data, &msg); err != nil {
			log.Println(err)
			break
		}
		if b, ok := ucaAction2Byte[msg.Action]; ok {
			d := ucaBytesForCommand(b, nil)
			log.Printf("uca writing %x", d)
			_, err := ucaPort.Write(d)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("unknown action: %s", msg.Action)
		}
	}
}

func acceptCoins() {
	var data []byte
	buf := make([]byte, 10)
	for {
		n, err := ucaPort.Read(buf)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}

		if n == 0 {
			log.Println("EOF")
			time.Sleep(2 * time.Second)
			continue
		}

		if buf[0] == 0x90 {
			data = []byte{}
		}
		data = append(data, buf[:n]...)

		if data != nil && len(data) > 1 && len(data) == int(data[1]) &&
			ucaGenerateChecksum(data[:len(data)-1]) == data[len(data)-1] {
			log.Printf("uca received %x", data)
			command := data[2]
			if s, ok := ucaByte2Status[command]; ok {
				var t *int
				if int(data[1]) > 5 {
					_t := int(data[3])
					t = &_t
				}
				b, _ := json.Marshal(baStatusReply{
					Action: "status",
					Status: s,
					Type:   t,
				})
				for client := range ucaClients {
					client.WriteMessage(websocket.TextMessage, b)
				}
			} else if command == 0x03 {
				info := data[3 : int(data[1])-3+1]
				version := string(info[:len(info)-2])
				b, _ := json.Marshal(ucaStatusReply{
					Action:  "status",
					Status:  "info",
					Version: &version,
				})
				for client := range ucaClients {
					client.WriteMessage(websocket.TextMessage, b)
				}
			}
			data = nil
		}
	}
}

func ucaBytesForCommand(command byte, data []byte) []byte {
	buf := append(append([]byte{0x90, byte(5 + len(data)), command}, data...), 0x03)
	return append(buf, ucaGenerateChecksum(buf))
}

func ucaBytesSum(buf []byte) int {
	sum := 0
	for i := 0; i < len(buf); i++ {
		sum += int(buf[i])
	}
	return sum
}

func ucaGenerateChecksum(buf []byte) byte {
	return byte(ucaBytesSum(buf) & 0x00FF)
}
