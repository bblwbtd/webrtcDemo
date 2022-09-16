package main

import (
	"flag"
	"stunDemo/pkg"
)

func main() {
	mode := flag.String("m", "i", "Initiator/recipient mode")
	flag.Parse()

	peer := pkg.NewPeer()
	if *mode == "i" {
		peer.GenerateOffer()
		peer.WaitForAnswer()
	} else if *mode == "r" {
		peer.GenerateAnswer()
	} else {
		panic("invalid mode")
	}

	select {}
}
