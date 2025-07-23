package main

import (
	"context"
	"dns-server/handlers"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var wg sync.WaitGroup
	rootCtx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer fmt.Println("Done listening")
		addr := ":53"
		pc, err := net.ListenPacket("udp", addr)
		if err != nil {
			panic(err)
		}
		defer pc.Close()
		log.Println("DNS server listening on", addr)

		buf := make([]byte, 512)
		for {
			select {
			case <-rootCtx.Done():
				return
			default:
				// Set a read timeout to allow periodic context checking
				pc.SetReadDeadline(time.Now().Add(1 * time.Second))
				n, clientAddr, err := pc.ReadFrom(buf)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// Timeout occurred, continue to check context
						continue
					}
					// Other error, log and continue
					log.Printf("Error reading from connection: %v", err)
					continue
				}
				wg.Add(1)
				go handlers.HandleDNSQuery(&wg, pc, clientAddr, buf[:n])
			}
		}
	}()

	go func() {
		defer wg.Done()

		sig := <-sigs
		log.Printf("Received Signal %v ", sig)
		cancel()
	}()

	wg.Wait()
}
