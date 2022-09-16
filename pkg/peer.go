package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc/v3"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type Peer struct {
	peerConnection *webrtc.PeerConnection
}

func NewPeer() *Peer {
	return &Peer{
		peerConnection: initPeerConnection(),
	}
}

func (p *Peer) GenerateAnswer() {
	generateAnswer(p.peerConnection)
}

func (p *Peer) GenerateOffer() {
	createDataChannel(p.peerConnection)
	generateOffer(p.peerConnection)
}

func (p *Peer) WaitForAnswer() {
	reader := bufio.NewReader(os.Stdin)
	_, _, _ = reader.ReadLine()

	bytes, err := os.ReadFile("answer.json")
	if err != nil {
		panic(err)
	}
	answer := webrtc.SessionDescription{}
	err = json.Unmarshal(bytes, &answer)
	if err != nil {
		panic(err)
	}

	err = p.peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}
}

func initPeerConnection() *webrtc.PeerConnection {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			panic("Peer Connection has gone to failed exiting")
		}
	})

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		d.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

			for range time.NewTicker(5 * time.Second).C {
				message := "Hi" + strconv.Itoa(rand.Int())
				fmt.Printf("Sending '%s'\n", message)

				sendErr := d.SendText(message)
				if sendErr != nil {
					panic(sendErr)
				}
			}
		})

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	return peerConnection
}

func createDataChannel(connection *webrtc.PeerConnection) *webrtc.DataChannel {
	channel, err := connection.CreateDataChannel("test", nil)
	if err != nil {
		panic(err)
	}

	channel.OnOpen(func() {
		fmt.Println("Data channel open")
		for range time.NewTicker(5 * time.Second).C {
			message := "Hi" + strconv.Itoa(rand.Int())
			fmt.Printf("Sending '%s'\n", message)

			sendErr := channel.SendText(message)
			if sendErr != nil {
				panic(sendErr)
			}
		}
	})

	channel.OnClose(func() {
		fmt.Println("Channel closed")
	})

	channel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Message from DataChannel '%s': '%s'\n", channel.Label(), string(msg.Data))
	})

	return channel
}

func generateOffer(connection *webrtc.PeerConnection) {
	offer, err := connection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(offer)
	if err != nil {
		panic(err)
	}

	err = connection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
	err = os.WriteFile("offer.json", bytes, 0777)
	if err != nil {
		panic(err)
	}
}

func generateAnswer(connection *webrtc.PeerConnection) {
	offer := webrtc.SessionDescription{}

	bytes, err := os.ReadFile("offer.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(bytes, &offer)
	if err != nil {
		panic(err)
	}

	err = connection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	answer, err := connection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	err = connection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(connection)

	<-gatherComplete

	marshal, _ := json.Marshal(*connection.LocalDescription())

	fmt.Println(string(marshal))
	err = os.WriteFile("answer.json", marshal, 0777)
	if err != nil {
		panic(err)
	}
}
