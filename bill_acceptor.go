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

	billAcceptorProtocol int
)

const (
	bapICT002U billAcceptorProtocol = iota
	bapICT104U
	bapICT106U
)

var (
	billAcceptorPort  serial.Port
	baCurrentProtocol billAcceptorProtocol

	baProtocols = map[string]billAcceptorProtocol{
		"ICT002U": bapICT002U,
		"ICT104U": bapICT104U,
		"ICT106U": bapICT106U,
	}
	baProtocolsRev = map[billAcceptorProtocol]string{
		bapICT002U: "ICT002U",
		bapICT104U: "ICT104U",
		bapICT106U: "ICT106U",
	}

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

	baAction2Bytes = map[billAcceptorProtocol]map[string][]byte{
		bapICT002U: {
			"enable":  {0x3e, 0x0c}, // reply status immediately
			"disable": {0x5e, 0x0c}, // reply status immediately
			"status":  {0x0c},
			"accept":  {0x02},
			"reject":  {0x0f},
		},
		bapICT104U: {
			"enable":  {0x3e, 0x0c}, // reply status immediately
			"disable": {0x5e, 0x0c}, // reply status immediately
			"reset":   {0x30},
			"status":  {0x0c},
			"accept":  {0x02},
			"reject":  {0x0f},
			"hold":    {0x18},
		},
		bapICT106U: {
			"enable":  {0x3e},
			"disable": {0x5e},
			"reset":   {0x30},
			"status":  {0x0c},
			"accept":  {0x02},
			"reject":  {0x0f},
			"hold":    {0x18},
			"info":    {0x5b},
		},
	}

	lastType *int
)

func startBillAcceptor(device, protocol string) {
	if proto, ok := baProtocols[protocol]; ok {
		baCurrentProtocol = proto
	} else {
		log.Fatalln("fatal: unsupported bill acceptor protocol", protocol)
	}
	var err error
	billAcceptorPort, err = serial.Open(device, &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.EvenParity,
		StopBits: serial.OneStopBit,
	})
	if err == nil {
		log.Println("opened bill acceptor device", device, "using protocol", protocol)
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

	actions := []string{}
	for action := range baAction2Bytes[baCurrentProtocol] {
		actions = append(actions, action)
	}
	b, _ := json.Marshal(struct {
		Action           string   `json:"action"`
		Protocol         string   `json:"protocol"`
		SupportedActions []string `json:"supported_actions"`
	}{"init", baProtocolsRev[baCurrentProtocol], actions})
	c.WriteMessage(websocket.TextMessage, b)

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
		if b, ok := baAction2Bytes[baCurrentProtocol][msg.Action]; ok {
			log.Printf("bill acceptor writing %x", b)
			_, err := billAcceptorPort.Write(b)
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
			} else if (baCurrentProtocol == bapICT002U || baCurrentProtocol == bapICT104U) && len(data) > 1 && data[0] == 0x81 {
				t := int(data[1]) - int(0x40)
				lastType = &t
				baReplyStatus("validated")
				data = nil
				continue
			} else if baCurrentProtocol == bapICT106U && len(data) > 2 && bytes.Equal(data[:2], []byte{0x81, 0x8f}) {
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
