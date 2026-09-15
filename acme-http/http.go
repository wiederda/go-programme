package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const http01PathPrefix = "/.well-known/acme-challenge/"

// HTTP01KeyAuthorization erzeugt den Wert, den Let's Encrypt
// beim HTTP-01-Challenge-Aufruf erwartet:
//
//	<token>.<JWK-thumbprint>
func HTTP01KeyAuthorization(token string, thumbprint string) string {
	return token + "." + thumbprint
}

// FindHTTP01Challenge sucht die HTTP-01-Challenge in einer
// ACME Authorization.
func FindHTTP01Challenge(authz *Authorization) (*Challenge, error) {
	if authz == nil {
		return nil, errors.New(
			"ACME Authorization ist nil",
		)
	}

	for i := range authz.Challenges {
		if authz.Challenges[i].Type == "http-01" {
			return &authz.Challenges[i], nil
		}
	}

	return nil, errors.New(
		"Authorization enthält keine HTTP-01-Challenge",
	)
}

// HTTPChallengeServer stellt genau eine HTTP-01-Challenge bereit.
type HTTPChallengeServer struct {
	server   *http.Server
	listener net.Listener
	errCh    chan error
	token    string
	keyAuth  string
}

// NewHTTPChallengeServer erzeugt einen HTTP-01-Server.
//
// listen kann zum Beispiel sein:
//
//	0.0.0.0:80
//	:80
//	127.0.0.1:8080
func NewHTTPChallengeServer(
	listen string,
	token string,
	keyAuthorization string,
) (*HTTPChallengeServer, error) {

	if token == "" {
		return nil, errors.New(
			"HTTP-01 Token ist leer",
		)
	}

	if keyAuthorization == "" {
		return nil, errors.New(
			"HTTP-01 KeyAuthorization ist leer",
		)
	}

	mux := http.NewServeMux()

	s := &HTTPChallengeServer{
		token:   token,
		keyAuth: keyAuthorization,
		errCh:   make(chan error, 1),
	}

	expectedPath := http01PathPrefix + token

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet &&
			r.Method != http.MethodHead {
			w.Header().Set(
				"Allow",
				http.MethodGet+", "+http.MethodHead,
			)
			http.Error(
				w,
				"Method Not Allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}

		w.Header().Set(
			"Content-Type",
			"text/plain; charset=utf-8",
		)
		w.Header().Set(
			"Cache-Control",
			"no-store",
		)

		if r.Method == http.MethodHead {
			return
		}

		_, _ = io.WriteString(w, keyAuthorization)
	})

	s.server = &http.Server{
		Addr:              listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, fmt.Errorf(
			"HTTP-01 Server konnte %s nicht öffnen: %w",
			listen,
			err,
		)
	}

	s.listener = listener

	go func() {
		err := s.server.Serve(listener)

		if err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			s.errCh <- err
		}
	}()

	return s, nil
}

// Address liefert die tatsächlich verwendete Listener-Adresse.
func (s *HTTPChallengeServer) Address() string {
	if s == nil || s.listener == nil {
		return ""
	}

	return s.listener.Addr().String()
}

// Stop beendet den HTTP-01-Server.
func (s *HTTPChallengeServer) Stop(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

// CheckError prüft, ob der HTTP-Server während des Betriebs
// einen unerwarteten Fehler gemeldet hat.
func (s *HTTPChallengeServer) CheckError() error {
	if s == nil {
		return nil
	}

	select {
	case err := <-s.errCh:
		return fmt.Errorf(
			"HTTP-01 Server wurde beendet: %w",
			err,
		)

	default:
		return nil
	}
}
