package main

import (
	"context"
	"dns-server/internal/constants"
	"dns-server/internal/handlers"
	"dns-server/internal/logger"
	"dns-server/internal/manager"
	"dns-server/internal/server"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func serverInit() {
	var err error
	logger.NewLogger(
		logger.WithLogFilePath("logs"),
		logger.WithLevel("info"),
	)

	constants.Redis, err = manager.NewRedisManager()
	if err != nil {
		log.Error().Msgf("Error while initialising Redis -> %v", err)
	}

	constants.ContextManager = manager.NewContextManager()
	handlers.LoadRedisContext()
}

func serverClose() {
	if constants.Redis != nil {
		constants.Redis.Close()
	}
}

func main() {
	serverInit()
	defer serverClose()

	var wg sync.WaitGroup
	rootCtx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	// Start DNS server
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer fmt.Println("Done listening")
		addr := ":53"
		pc, err := net.ListenPacket("udp", addr)
		if err != nil {
			panic(err)
		}
		defer pc.Close()
		log.Info().Msgf("DNS server listening on %s", addr)

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
					log.Error().Msgf("Error reading from connection: %v", err)
					continue
				}
				wg.Add(1)
				go handlers.HandleDNSQuery(rootCtx, &wg, pc, clientAddr, buf[:n])
			}
		}
	}()

	// Start API server
	go func() {
		srv := server.NewServer()
		if err := srv.StartBackend(); err != nil && err != http.ErrServerClosed {
			log.Panic().Msgf("Failed to start server at addr %s -> error %v", srv.GetAddr(), err)
		}
	}()

	// if gin.Mode() == gin.ReleaseMode {
	go func() {
		if err := server.StartFrontend(); err != nil && err != http.ErrServerClosed {
			log.Panic().Msgf("Failed to start server at addr 3002 -> error %v", err)
		}
	}()
	// }

	// Signal handler
	wg.Add(1)
	go func() {
		defer wg.Done()

		sig := <-sigs
		log.Printf("Received Signal %v ", sig)
		cancel()
	}()

	wg.Wait()
}
