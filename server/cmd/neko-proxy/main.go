package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	nekoproxy "github.com/m1k1o/neko/server/internal/proxy"
)

const defaultListenAddress = "127.0.0.1:18080"

const (
	defaultHealthListenAddress = "127.0.0.1:18081"
	healthCheckAttempts        = 3
	healthCheckTimeout         = 10 * time.Second
	healthCheckInterval        = 30 * time.Second
)

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
	healthListen := os.Getenv("NEKO_CHROMIUM_PROXY_HEALTH_LISTEN")
	if healthListen == "" {
		healthListen = defaultHealthListenAddress
	}
	healthListen, err = nekoproxy.ListenAddress(healthListen)
	if err != nil {
		return err
	}
	healthTarget, err := nekoproxy.TargetAddress(os.Getenv("NEKO_CHROMIUM_PROXY_HEALTHCHECK_TARGET"))
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
	monitor, err := nekoproxy.NewHealthMonitor(agent, endpoint, healthTarget)
	if err != nil {
		return err
	}
	if err := initialHealthCheck(monitor); err != nil {
		return err
	}

	proxyListener, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("listen for Chromium proxy traffic: %w", err)
	}
	defer proxyListener.Close()
	healthListener, err := net.Listen("tcp", healthListen)
	if err != nil {
		return fmt.Errorf("listen for Chromium proxy health checks: %w", err)
	}
	defer healthListener.Close()

	httpServer := &http.Server{
		Handler:           agent,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	healthServer := &http.Server{
		Handler:           monitor,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go monitor.Run(shutdownContext, healthCheckInterval, healthCheckTimeout)
	go func() {
		<-shutdownContext.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
		_ = healthServer.Shutdown(ctx)
	}()

	log.Printf("Chromium proxy agent listening on %s with upstream %s; health endpoint on %s", listen, endpoint.RedactedServer(), healthListen)
	errorsChannel := make(chan error, 2)
	go func() { errorsChannel <- httpServer.Serve(proxyListener) }()
	go func() { errorsChannel <- healthServer.Serve(healthListener) }()
	err = <-errorsChannel
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func initialHealthCheck(monitor *nekoproxy.HealthMonitor) error {
	for attempt := 1; attempt <= healthCheckAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
		err := monitor.Check(ctx)
		cancel()
		if err == nil {
			return nil
		}
		if attempt < healthCheckAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return fmt.Errorf("Chromium proxy startup health check failed: %s", monitor.Snapshot().Reason)
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
