package main

// Bluetooth Service for Blockie Talkie
// UART over bluetooth and creates a UNIX socket for inter-app communication

// Intended to run on a linux platform

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"tinygo.org/x/bluetooth"
)

var (
	serviceUUID = bluetooth.ServiceUUIDNordicUART
	rxUUID      = bluetooth.CharacteristicUUIDUARTRX
	txUUID      = bluetooth.CharacteristicUUIDUARTTX
)

func main() {
	println("starting")
	adapter := bluetooth.DefaultAdapter
	must("enable BLE stack", adapter.Enable())
	adv := adapter.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Clock", // Nordic UART Service
		ServiceUUIDs: []bluetooth.UUID{serviceUUID},
	}))
	must("start adv", adv.Start())

	var rxChar bluetooth.Characteristic
	var txChar bluetooth.Characteristic
	var name = []byte("User")
	must("add service", adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &rxChar,
				UUID:   rxUUID,
				Flags:  bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					name, value = decodeMsg(value, name)
					// Send over socket 'echo'
					// sendOverSocket(value)
					// TODO: Make some kind of sent confirmation for device
					txChar.Write(value)
					fmt.Println("Sent Message:", string(value))
				},
			},
			{
				Handle: &txChar,
				UUID:   txUUID,
				Flags:  bluetooth.CharacteristicNotifyPermission | bluetooth.CharacteristicReadPermission,
			},
		},
	}))

	// Setup Unix Socket
	socketPath := "/tmp/blockie_talkie_comm"

	// Ensure the socket file does not exist before listening
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on:", socketPath)
	for {
		// unix socket connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleSocketConnection(conn)
	}
}

func decodeMsg(value []byte, nameForMsg []byte) ([]byte, []byte) {
	// TODO: Brainstorm other ideas on how to send commands
	// message commands utilizing the '=' as a separator
	decodedMsg := string(value)
	// change name on getting name={your name}
	if strings.Contains(decodedMsg, "name=") {
		nameVal := strings.SplitAfter(decodedMsg, "=")
		if len(nameVal) > 1 {
			nameForMsg = []byte(nameVal[1])
		}
	}
	name := []byte(nameForMsg)
	separator := []byte(": ")
	nameWSep := append(name, separator...)
	newMsg := append(nameWSep, value...)

	return name, newMsg
}

func getBluetoothData() ([]byte, error) {
	// Example: Using `bluetoothctl` just keeping this here for ref on a command
	cmd := exec.Command("bluetoothctl", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing bluetoothctl: %v", err)
	}
	return output, nil
}

func handleSocketConnection(conn net.Conn) {
	// Handle UNIX socket messages
	// This will be where the incoming messages will be forwarded to bluetooth
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading:", err)
			}
			return
		}
		fmt.Printf("Received Unix Socket: %s\n", string(buf[:n]))
		_, err = conn.Write([]byte("Message received\n"))
		if err != nil {
			fmt.Println("Error writing:", err)
			return
		}
	}
}

func readOverSocket() {
	// Unix socket
	socketPath := "/tmp/blockie_talkie_comm"
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on:", socketPath)
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting:", err)
	}
	go handleSocketConnection(conn)
}

func sendOverSocket(msg []byte) {
	// Unix socket
	socketPath := "/tmp/blockie_talkie_comm"
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer conn.Close()
	message := msg
	_, err = conn.Write(message)
	if err != nil {
		fmt.Println("Error writing:", err)
		os.Exit(1)
	}
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		os.Exit(1)
	}
	fmt.Println("Received Send:", string(buffer[:n]))
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
