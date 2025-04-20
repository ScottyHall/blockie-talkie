package main

// UNIX Socket Send for testing/inspiration/copying

// Intended to run on a linux platform. Run on same system as bluetooth service

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
)

func getBluetoothData() ([]byte, error) {
	cmd := exec.Command("bluetoothctl", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing bluetoothctl: %v", err)
	}
	return output, nil
}

func main() {
	// Socket needs to exist, otherwise this will error, assumes you are running blockie_blue_service
	socketPath := "/tmp/blockie_talkie_comm"
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		fmt.Println("Error connecting, check if blockie_blue_service is running. ERROR:", err)
		os.Exit(1)
	}
	defer conn.Close()

	message := []byte("Sup from the other side of the socket")
	_, err = conn.Write(message)
	if err != nil {
		fmt.Println("Error writing:", err)
		os.Exit(1)
	}
	fmt.Println("Sent:", string(message))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		os.Exit(1)
	}
	fmt.Println("Received:", string(buffer[:n]))

	message2 := []byte("Anotha one")

	_, err = conn.Write(message2)

	if err != nil {
		fmt.Println("Error writing second message", err)
		os.Exit(1)
	}
	fmt.Println("Sent:", string(message2))

	n2, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Error reading second message", err)
		os.Exit(1)
	}

	fmt.Println("Received", string(buffer[:n2]))

	// Get Bluetooth data
	bluetoothData, err := getBluetoothData()
	if err != nil {
		log.Fatalf("Error getting Bluetooth data: %v", err)
	}
	blueMsg := []byte(bluetoothData)
	_, err = conn.Write(blueMsg)

	n3, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Error reading second message", err)
		os.Exit(1)
	}

	fmt.Println("Received", string(buffer[:n3]))
}
