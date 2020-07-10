package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.bug.st/serial"
)

type (
	mhActionMessage struct {
		Action string `json:"action"`
		Value  int    `json:"value"`
	}

	mhStatusReply struct {
		Action string `json:"action"`
		Status string `json:"status"`
	}
)

const (
	statusMotorProblem = 1 << iota
	statusHopperLowLevelDetected
	_
	statusPrismSensorFailure
	statusShaftSensorFailure
	statusBusy
)

var (
	miniHopperPort serial.Port

	mhUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mhClients = map[*websocket.Conn]bool{}

	mhByte2Status = map[byte]string{
		0x04: "idle",
		0x07: "payout-once",
		0x08: "payout-finished",
		0xbb: "busy",
	}

	mhAction2Byte = map[string]byte{
		"payout":         0x10,
		"status":         0x11,
		"reset":          0x12,
		"remaining":      0x13,
		"payout+message": 0x14,
		"empty+message":  0x15,
		"stop":           0x17,
	}
)

func startMiniHopper(device string) {
	var err error
	miniHopperPort, err = serial.Open(device, &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.EvenParity,
		StopBits: serial.OneStopBit,
	})
	if err == nil {
		log.Println("opened mini hopper device", device)
	} else {
		log.Fatal(err)
	}
	go giveChange()
	http.HandleFunc("/ict/mini-hopper", mhServer)
}

func mhServer(w http.ResponseWriter, r *http.Request) {
	c, err := mhUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	defer func() {
		delete(mhClients, c)
		c.Close()
	}()
	mhClients[c] = true
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Println("read error:", err)
			}
			break
		}
		var msg mhActionMessage
		if err = json.Unmarshal(data, &msg); err != nil {
			log.Println(err)
			break
		}
		if b, ok := mhAction2Byte[msg.Action]; ok {
			d := mhBytesForCommand(b, byte(msg.Value))
			log.Printf("mini hopper writing %x", d)
			_, err := miniHopperPort.Write(d)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("unknown action: %s", msg.Action)
		}
	}
}

func giveChange() {
	var data []byte
	buf := make([]byte, 10)
	for {
		n, err := miniHopperPort.Read(buf)
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

		if buf[0] == 0x05 && buf[1] == 0x01 {
			data = []byte{}
		}
		data = append(data, buf[:n]...)

		if data != nil && len(data) > 3 && mhGenerateChecksum(data[:len(data)-1]) == data[len(data)-1] {
			log.Printf("mini hopper received %x", data)
			command := data[3]
			s, ok := mhByte2Status[command]
			if !ok {
				continue
			}
			if command == 0x04 {
				ss := []string{}
				if data[4]&statusMotorProblem != 0 {
					ss = append(ss, "motor_problem")
				}
				if data[4]&statusHopperLowLevelDetected != 0 {
					ss = append(ss, "hopper_low_level_detected")
				}
				if data[4]&statusPrismSensorFailure != 0 {
					ss = append(ss, "prism_sensor_failure")
				}
				if data[4]&statusShaftSensorFailure != 0 {
					ss = append(ss, "shaft_sensor_failure")
				}
				if data[4]&statusBusy != 0 {
					ss = append(ss, "busy")
				}
				if len(ss) > 0 {
					s = strings.Join(ss, ",")
				}
			}
			b, _ := json.Marshal(mhStatusReply{
				Action: "status",
				Status: s,
			})
			for client := range mhClients {
				client.WriteMessage(websocket.TextMessage, b)
			}
		}
	}
}

func mhBytesForCommand(command, data byte) []byte {
	buf := []byte{0x05, 0x10, 0x00, command, data}
	return append(buf, mhGenerateChecksum(buf))
}

func mhBytesSum(buf []byte) int {
	sum := 0
	for i := 0; i < len(buf); i++ {
		sum += int(buf[i])
	}
	return sum
}

func mhGenerateChecksum(buf []byte) byte {
	return byte(ucaBytesSum(buf) & 0x00FF)
}
