package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	flagNats    string
	flagTLSKey  string
	flagTLSCert string
	flagTLSCA   string
)

func main() {
	flag.StringVar(&flagNats, "nats", "nats://127.0.0.1:4222/", "Nats to connect to")
	flag.StringVar(&flagTLSKey, "key", "", "TLS Client Key")
	flag.StringVar(&flagTLSCert, "cert", "", "TLS Client Cert")
	flag.StringVar(&flagTLSCA, "ca", "", "TLS CA Cert")
	flag.Parse()
	siphonNats()
}

func makeTLSConfig() (*tls.Config, error) {
	c := new(tls.Config)

	if len(flagTLSCA) != 0 {
		pemBytes, err := ioutil.ReadFile(flagTLSCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read crypto ca cert: %w", err)
		}

		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return nil, fmt.Errorf("data was not pem: %w", err)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse crypto ca cert: %w", err)
		}

		c.RootCAs = x509.NewCertPool()
		c.RootCAs.AddCert(cert)
	}

	cert, err := tls.LoadX509KeyPair(flagTLSCert, flagTLSKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load crypto certificate and key: %w", err)
	}

	c.Certificates = append(c.Certificates, cert)
	return c, nil
}

func siphonNats() {
	var conn *nats.Conn
	var err error

	if len(flagTLSCert) != 0 || len(flagTLSKey) != 0 || len(flagTLSCA) != 0 {
		config, err := makeTLSConfig()
		if err != nil {
			log.Println("failed to create tls config:", err)
			os.Exit(1)
		}

		log.Println("[tls] connecting to:", flagNats)
		conn, err = nats.Connect(flagNats, nats.Secure(config))
	} else {
		log.Println("connecting to:", flagNats)
		conn, err = nats.Connect(flagNats)
	}

	if err != nil {
		log.Println("could not connect to nats:", err)
		os.Exit(1)
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
					break
				} else {
					log.Printf("[%s] %s", msg.Subject, msg.Data)
				}
			}
			wg.Done()
		}(sub)
	}

	wg.Wait()
}
