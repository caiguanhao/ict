package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

type (
	baActionMessage struct {
		Action string `json:"action"`
	}

	baStatusReply struct {
		Action string `json:"action"`
		Status string `json:"status"`
		Type   *int   `json:"type"`
	}
)

var (
	billAcceptorPort serial.Port

	baUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	baClients = map[*websocket.Conn]bool{}

	baByte2Status = map[byte]string{
		0x20: "motor_failure",
		0x21: "checksum_error",
		0x22: "bill_jam",
		0x23: "bill_remove",
		0x24: "stacker_open",
		0x25: "sensor_problem",
		0x27: "bill_fish",
		0x28: "stacker_problem",
		0x29: "bill_reject",
		0x2a: "invalid_command",
		0x3e: "enable",
		0x5e: "disable",
		0x71: "busy",
		0xa1: "power_supply",
		0x10: "done",
		0x11: "reject",
	}

	baAction2Byte = map[string]byte{
		"enable":  0x3e,
		"disable": 0x5e,
		"reset":   0x30,
		"status":  0x0c,
		"accept":  0x02,
		"reject":  0x0f,
		"hold":    0x18,
		"info":    0x5b,
	}

	lastType *int
)

func startBillAcceptor(device string) {
	var err error
	billAcceptorPort, err = serial.Open(device, &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.EvenParity,
		StopBits: serial.OneStopBit,
	})
	if err == nil {
		log.Println("opened bill acceptor device", device)
	} else {
		log.Fatal(err)
	}
	go acceptBills()
	http.HandleFunc("/ict/bill-acceptor", baServer)
}

func baServer(w http.ResponseWriter, r *http.Request) {
	c, err := baUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	defer func() {
		delete(baClients, c)
		c.Close()
	}()
	baClients[c] = true
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Println("read error:", err)
			}
			break
		}
		var msg baActionMessage
		if err = json.Unmarshal(data, &msg); err != nil {
			log.Println(err)
			break
		}
		if b, ok := baAction2Byte[msg.Action]; ok {
			log.Printf("bill acceptor writing %x", []byte{b})
			_, err := billAcceptorPort.Write([]byte{b})
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("unknown action: %s", msg.Action)
		}
	}
}

func acceptBills() {
	var data []byte
	buf := make([]byte, 3)
	for {
		n, err := billAcceptorPort.Read(buf)
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

		if s, ok := baByte2Status[buf[0]]; ok {
			log.Printf("bill acceptor received %x", buf[:n])
			baReplyStatus(s)
			data = nil
			lastType = nil
			continue
		}

		if buf[0] == 0x80 || buf[0] == 0x81 {
			data = []byte{}
		}
		data = append(data, buf[:n]...)

		if data != nil {
			log.Printf("bill acceptor received %x", data)
			if bytes.Equal(data[:2], []byte{0x80, 0x8f}) {
				_, err := billAcceptorPort.Write([]byte{0x02})
				if err == nil {
					baReplyStatus("reset")
				} else {
					log.Println(err)
					time.Sleep(2 * time.Second)
				}
				data = nil
				lastType = nil
				continue
			} else if len(data) > 2 && bytes.Equal(data[:2], []byte{0x81, 0x8f}) {
				t := int(data[2]) - int(0x40)
				lastType = &t
				baReplyStatus("validated")
				data = nil
				continue
			}
		}
	}
}

func baReplyStatus(status string) {
	b, _ := json.Marshal(baStatusReply{
		Action: "status",
		Status: status,
		Type:   lastType,
	})
	for client := range baClients {
		client.WriteMessage(websocket.TextMessage, b)
	}
}
