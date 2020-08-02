package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pfring"
	godpi "github.com/mushorg/go-dpi"
	"github.com/mushorg/go-dpi/types"
	"os"
	"os/signal"
)

var device = flag.String("device", "", "Device to watch for packets")

func main() {
	var (
		count, idCount int
		protoCounts    map[types.Protocol]int
		packetChannel  <-chan gopacket.Packet
		err            error
	)

	protoCounts = make(map[types.Protocol]int)

	flag.Parse()

	if *device == "" {
		fmt.Println("Please specify a interface with -device option")
		return
	}

	handle, err := pfring.NewRing(string(*device), 65536, pfring.FlagPromisc)
	if err != nil {
		fmt.Println("Error opening device:", err)
		return
	}

	if err = handle.SetSocketMode(pfring.ReadOnly); err != nil {
		fmt.Println("pfring SetSocketMode error:", err)
		return
	}

	if err = handle.Enable(); err != nil {
		fmt.Println("pfring enable error:", err)
		return
	}

	defer handle.Close()

	packetChannel = gopacket.NewPacketSource(handle, layers.LayerTypeEthernet).Packets()

	initErrs := godpi.Initialize()
	if len(initErrs) != 0 {
		for _, err := range initErrs {
			fmt.Println(err)
		}
		return
	}

	defer func() {
		godpi.Destroy()
		fmt.Println()
		fmt.Println("Number of packets:", count)
		fmt.Println("Number of packets identified:", idCount)
		fmt.Println("Protocols identified:\n", protoCounts)
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	intSignal := false

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	count = 0
	for packet := range packetChannel {
		fmt.Printf("Packet #%d: ", count+1)
		flow, isNew := godpi.GetPacketFlow(packet)
		result := godpi.ClassifyFlow(flow)
		if result.Protocol != types.Unknown {
			fmt.Print(result)
			idCount++
			protoCounts[result.Protocol]++
		} else {
			fmt.Print("Could not identify")
		}
		if isNew {
			fmt.Println(" (new flow)")
		} else {
			fmt.Println()
		}

		select {
		case <-signalChannel:
			fmt.Println("Received interrupt signal")
			intSignal = true
		default:
		}
		if intSignal {
			break
		}
		count++
	}
}

