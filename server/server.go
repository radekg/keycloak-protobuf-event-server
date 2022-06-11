package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/radekg/keycloak-protobuf-event-server/config"
	"github.com/radekg/keycloak-protobuf-spi/gospi/eventlistener"
	"github.com/radekg/keycloak-protobuf-spi/gospi/shared"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	sync.Mutex

	config *config.ServerConfig
	logger hclog.Logger

	chanReady   chan struct{}
	chanStopped chan struct{}
	chanFailed  chan struct{}

	failedError error

	srv *grpc.Server

	wasStarted bool
	running    bool
}

func NewServer(cfg *config.ServerConfig, logger hclog.Logger) *Server {
	return &Server{
		config:      cfg,
		logger:      logger,
		chanFailed:  make(chan struct{}),
		chanReady:   make(chan struct{}),
		chanStopped: make(chan struct{}),
	}
}

func (s *Server) Start() {

	s.Lock()
	defer s.Unlock()

	if !s.wasStarted {

		s.wasStarted = true

		listener, err := net.Listen("tcp", s.config.BindHostPort)
		if err != nil {
			s.logger.Error("Failed to create TCP listener", "hostport", s.config.BindHostPort, "reason", err)
			s.failedError = err
			close(s.chanFailed)
			return
		}

		s.logger.Info("TCP listener created", "hostport", s.config.BindHostPort)

		if !s.config.NoTLS {

			s.logger.Info("Starting with TLS")

			certificate, err := tls.LoadX509KeyPair(s.config.TLSCertificateFilePath, s.config.TLSKeyFilePath)
			if err != nil {
				s.logger.Error("Failed to load server certificate or key",
					"cert-file-path", s.config.TLSCertificateFilePath,
					"key-file-path", s.config.TLSKeyFilePath,
					"reason", err)
				s.failedError = err
				close(s.chanFailed)
				return
			}

			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{certificate},
			}

			if s.config.TLSTrustedCertificatesFilePath != "" {
				certPool := x509.NewCertPool()
				ca, err := ioutil.ReadFile(s.config.TLSTrustedCertificatesFilePath)
				if err != nil {
					s.logger.Error("Failed to load trusted certificate",
						"trusted-cert-file-path", s.config.TLSTrustedCertificatesFilePath,
						"reason", err)
					s.failedError = err
					close(s.chanFailed)
					return
				}
				if ok := certPool.AppendCertsFromPEM(ca); !ok {
					s.logger.Error("Failed to append trusted certificate to the cert pool", "reason", err)
					s.failedError = err
					close(s.chanFailed)
					return
				}
				tlsConfig.ClientCAs = certPool
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			}

			creds := credentials.NewTLS(tlsConfig)
			s.srv = grpc.NewServer(grpc.Creds(creds))
		} else {

			s.logger.Warn("Starting without TLS, use TLS in production")
			s.srv = grpc.NewServer()

		}

		s.logger.Info("Registering seervice with the GRPC server")
		eventlistener.RegisterKeycloakEventServiceServer(s.srv, s)

		chanErr := make(chan struct{})
		go func() {
			if err := s.srv.Serve(listener); err != nil {
				s.logger.Error("Failed to serve", "reason", "error")
				s.failedError = err
				close(chanErr)
				close(s.chanFailed)
			}
		}()

		select {
		case <-chanErr:
		case <-time.After(100):
			s.logger.Info("GRPC server running")
			s.running = true
			close(s.chanReady)
		}

	} else {
		s.logger.Warn("Server was already started, can't start twice")
	}

}

// Stop stops the server, if the server is started.
func (s *Server) Stop() {

	s.Lock()
	defer s.Unlock()

	if s.running {

		s.logger.Info("Attempting graceful stop")

		chanSignal := make(chan struct{})
		go func() {
			s.srv.GracefulStop()
			close(chanSignal)
		}()

		select {
		case <-chanSignal:
			s.logger.Info("Stopped gracefully")
		case <-time.After(time.Millisecond * time.Duration(s.config.GracefulStopTimeoutMillis)):
			s.logger.Warn("Failed to stop gracefully within timeout, forceful stop")
			s.srv.Stop()
		}

		s.logger.Info("Stopped")

		s.running = false
		close(s.chanStopped)

	} else {
		s.logger.Warn("Server not running")
	}

}

// ReadyNotify returns a channel that will be closed when the server is ready to serve client requests.
func (s *Server) ReadyNotify() <-chan struct{} {
	return s.chanReady
}

// FailedNotify returns a channel that will be closed when the server has failed to start.
func (s *Server) FailedNotify() <-chan struct{} {
	return s.chanFailed
}

// StoppedNotify returns a channel that will be closed when the server has stopped.
func (s *Server) StoppedNotify() <-chan struct{} {
	return s.chanStopped
}

// StartFailureReason returns an error which caused the server not to start.
func (s *Server) StartFailureReason() error {
	return s.failedError
}

// -- server implementation

// OnAdminEvent is the implementation of the OnAdminEvent protobuf service.
func (s *Server) OnAdminEvent(ctx context.Context, request *eventlistener.AdminEventRequest) (*shared.Empty, error) {
	s.logger.Info("OnAdminEvent", "admin-event", request.GetAdminEvent())
	return &shared.Empty{}, nil
}

// OnEvent is the implementation of the OnEvent protobuf service.
func (s *Server) OnEvent(ctx context.Context, request *eventlistener.EventRequest) (*shared.Empty, error) {
	s.logger.Info("OnEvent", "event", request.GetEvent())
	return &shared.Empty{}, nil
}
