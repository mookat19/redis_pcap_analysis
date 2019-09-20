package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	pcapFile string = "dump.pcap"
	handle   *pcap.Handle
	err      error
)

type pktinfo struct {
	payload   string
	offset    int
	timestamp time.Time
}

type pdata struct {
	count int
	pkts  []pktinfo
}

func main() {
	// Open file instead of device
	handle, err = pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packets := make(map[uint32]*pdata)

	// Set filter
	var filter string = "tcp"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	packetCount := 1

	for packet := range packetSource.Packets() {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if len(tcp.Payload) > 1 {
				_, ok := packets[tcp.Seq]
				if ok {
					packets[tcp.Seq].count++
					packets[tcp.Seq].pkts = append(packets[tcp.Seq].pkts, pktinfo{offset: packetCount, payload: string(tcp.Payload), timestamp: packet.Metadata().Timestamp})
				} else {
					packets[tcp.Seq] = &pdata{count: 1, pkts: append([]pktinfo{}, pktinfo{offset: packetCount, payload: string(tcp.Payload), timestamp: packet.Metadata().Timestamp})}
				}
			}

		}
		packetCount++
	}

	for key, value := range packets {
		if value.count > 1 {
			if value.pkts[0].payload == value.pkts[len(value.pkts)-1].payload {
				fmt.Printf("sequence: %d, time_diff: %d ms\n", key, value.pkts[len(value.pkts)-1].timestamp.Sub(value.pkts[0].timestamp).Milliseconds())
			}

		}
	}
}