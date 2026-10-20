package server // import "admin-console/internal/server"
import (
	"admin-console/internal/config"
	"admin-console/internal/storage"
	"admin-console/internal/worker"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func StartWebServer(store *storage.Storage, pool *worker.Pool) ([]*http.Server, func()) {
	var servers []*http.Server

	targets := determineListenTargets(config.Opts.ListenAddr(), "", "", "")

	for _, t := range targets {
		srv := &http.Server{
			Addr:              t.address,
			ReadTimeout:       config.Opts.HTTPServerTimeout(),
			WriteTimeout:      config.Opts.HTTPServerTimeout(),
			IdleTimeout:       config.Opts.HTTPServerTimeout(),
			ReadHeaderTimeout: config.Opts.HTTPServerTimeout(),
			Handler:           newRouter(store, pool),
		}
		switch t.mode {
		// case modeSystemd:
		// 	startSystemdSocketServer(srv)
		// case modeUnixSocket:
		// 	startUnixSocketServer(srv, t.address)
		// case modeUnixSocketTLS:
		// 	startUnixSocketTLSServer(srv, t.address, certLoader)
		// case modeAutocertTLS:
		// 	startAutoCertTLSServer(srv, autocertTLSConfig)
		// case modeTLS:
		// 	startTLSServer(srv, certLoader)
		default:
			startHTTPServer(srv)
		}

		servers = append(servers, srv)
	}

	certReloadFn := func() {}

	return servers, certReloadFn

}

type listenerMode int

const (
	modeHTTP listenerMode = iota
	modeTLS
	modeAutocertTLS
	modeUnixSocket
	modeUnixSocketTLS
	modeSystemd
)

type listenTarget struct {
	address string
	mode    listenerMode
}

func determineListenTargets(addresses []string, certDomain, certFile, keyFile string) []listenTarget {
	hasCertFiles := certFile != "" && keyFile != ""
	hasAutocert := certDomain != ""
	var targets []listenTarget

	for i, addr := range addresses {
		isUnix := strings.HasPrefix(addr, "/")
		switch {
		case isUnix && hasCertFiles:
			targets = append(targets, listenTarget{address: addr, mode: modeUnixSocketTLS})
		case isUnix:
			targets = append(targets, listenTarget{address: addr, mode: modeUnixSocket})
		case hasAutocert && (addr == ":https" || (i == 0 && strings.Contains(addr, ":"))):
			targets = append(targets, listenTarget{address: addr, mode: modeAutocertTLS})
		case hasCertFiles:
			targets = append(targets, listenTarget{address: addr, mode: modeTLS})
		default:
			targets = append(targets, listenTarget{address: addr, mode: modeHTTP})
		}
	}

	return targets
}

func startHTTPServer(server *http.Server) {
	go func() {
		slog.Info("Starting HTTP server",
			slog.String("listen_address", server.Addr),
		)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			printErrorAndExit("HTTP server failed to start on %s: %v", server.Addr, err)
		}
	}()
}

func printErrorAndExit(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	slog.Error(message)
	fmt.Fprintf(os.Stderr, "%v\n", message)
	os.Exit(1)
}
