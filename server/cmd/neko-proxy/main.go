package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	nekoproxy "github.com/m1k1o/neko/server/internal/proxy"
)

const defaultListenAddress = "127.0.0.1:18080"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	server := os.Getenv("NEKO_CHROMIUM_PROXY_SERVER")
	if server == "" {
		log.Print("Chromium proxy agent disabled: no upstream server configured")
		return nil
	}

	listen := os.Getenv("NEKO_CHROMIUM_PROXY_AGENT")
	if listen == "" {
		listen = defaultListenAddress
	}
	listen, err := nekoproxy.ListenAddress(listen)
	if err != nil {
		return err
	}

	password, err := readPassword(os.Getenv("NEKO_CHROMIUM_PROXY_PASSWORD_FILE"))
	if err != nil {
		return err
	}
	var bypass []string
	if value := os.Getenv("NEKO_CHROMIUM_PROXY_BYPASS_LIST"); value != "" {
		bypass = strings.Split(value, ";")
	}
	config := nekoproxy.Config{
		Server:   server,
		Username: os.Getenv("NEKO_CHROMIUM_PROXY_USERNAME"),
		Password: password,
		Bypass:   bypass,
	}
	agent, err := nekoproxy.NewAgent(config)
	if err != nil {
		return fmt.Errorf("invalid Chromium proxy configuration: %w", err)
	}
	endpoint, _ := config.Parse()

	httpServer := &http.Server{
		Addr:              listen,
		Handler:           agent,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownContext.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}()

	log.Printf("Chromium proxy agent listening on %s with upstream %s", listen, endpoint.RedactedServer())
	err = httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func readPassword(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read Chromium proxy password file: %w", err)
	}
	password := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	if password == "" {
		return "", errors.New("Chromium proxy password file is empty")
	}
	return password, nil
}
