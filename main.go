package main

import (
	"flag"
	"log"
	"sync"
	"time"

	"github.com/apcera/nats"
)

var (
	flagNats = flag.String("nats", "nats://127.0.0.1:4222/", "Nats to connect to")
)

func main() {
	flag.Parse()
	siphonNats()
}

func siphonNats() {
	log.Println("connecting to:", *flagNats)
	conn, err := nats.Connect(*flagNats)
	if err != nil {
		log.Println("could not connect to nats:", err)
		return
	}
	defer conn.Close()

	subs := []string{">"}
	if len(flag.Args()) > 0 {
		subs = flag.Args()
	}

	wg := &sync.WaitGroup{}

	for _, s := range subs {
		log.Println("subscribing to:", s)
		sub, err := conn.SubscribeSync(s)
		if err != nil {
			log.Println("failed to subscribe to nats:", err)
			return
		}

		wg.Add(1)
		go func(n *nats.Subscription) {
			for {
				if msg, err := n.NextMsg(1 * time.Hour); err != nil {
					log.Println("error getting nats msg:", err)
				} else {
					log.Printf("[%s] %s", msg.Subject, msg.Data)
				}
			}
			wg.Done()
		}(sub)
	}

	wg.Wait()
}
