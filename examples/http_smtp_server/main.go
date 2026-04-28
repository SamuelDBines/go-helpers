package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SamuelDBines/go-helpers/pkg/certs"
	"github.com/SamuelDBines/go-helpers/pkg/httpserver"
	"github.com/SamuelDBines/go-helpers/pkg/smtp"
)

type message struct {
	From       string    `json:"from"`
	To         []string  `json:"to"`
	Body       string    `json:"body"`
	ReceivedAt time.Time `json:"received_at"`
}

type mailbox struct {
	mu       sync.Mutex
	messages []message
}

func (m *mailbox) add(msg message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *mailbox) all() []message {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]message, len(m.messages))
	copy(out, m.messages)
	return out
}

type backend struct {
	box *mailbox
}

func (b *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{box: b.box}, nil
}

type session struct {
	box  *mailbox
	from string
	to   []string
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	s.to = nil
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	s.box.add(message{
		From:       s.from,
		To:         append([]string(nil), s.to...),
		Body:       string(body),
		ReceivedAt: time.Now().UTC(),
	})
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tlsConfig, err := certs.ServerTLSConfig(certs.Config{
		CertFile:     "certs/dev-cert.pem",
		KeyFile:      "certs/dev-key.pem",
		Hosts:        []string{"localhost", "127.0.0.1", "::1"},
		Organization: "go-helpers examples",
	})
	if err != nil {
		log.Fatalf("load tls config: %v", err)
	}

	box := &mailbox{}

	mux := http.NewServeMux()
	httpserver.With(mux, "/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpserver.OK(w, map[string]any{
			"status": "ok",
			"http":   "https://localhost:8443",
			"smtp":   "localhost:2525 (STARTTLS available)",
		})
	}))
	httpserver.With(mux, "/messages", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		messages := box.all()
		httpserver.OK(w, map[string]any{
			"count":    len(messages),
			"messages": messages,
		})
	}))

	httpSrv := httpserver.NewServer(httpserver.Config{
		Port: 8443,
		Name: "go-helpers-example",
	}, mux)
	httpSrv.TLSConfig = tlsConfig
	httpSrv.ReadTimeout = 10 * time.Second
	httpSrv.WriteTimeout = 10 * time.Second
	httpSrv.IdleTimeout = 30 * time.Second

	smtpSrv := smtp.NewServer(&backend{box: box})
	smtpSrv.Addr = "127.0.0.1:2525"
	smtpSrv.Domain = "localhost"
	smtpSrv.TLSConfig = tlsConfig
	smtpSrv.ReadTimeout = 30 * time.Second
	smtpSrv.WriteTimeout = 30 * time.Second

	errCh := make(chan error, 2)

	go func() {
		log.Printf("HTTPS server listening on %s", httpSrv.Addr)
		if err := httpSrv.ListenAndServeTLS("certs/dev-cert.pem", "certs/dev-key.pem"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		log.Printf("SMTP server listening on %s", smtpSrv.Addr)
		if err := smtpSrv.ListenAndServe(); err != nil && !errors.Is(err, smtp.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")
	case err := <-errCh:
		log.Fatalf("server failed: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("http shutdown: %v", err)
	}
	if err := smtpSrv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("smtp shutdown: %v", err)
	}
}
