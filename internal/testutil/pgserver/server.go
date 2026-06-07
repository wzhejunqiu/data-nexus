// Package pgserver starts an in-process PostgreSQL server for integration tests.
package pgserver

import (
	"fmt"
	"io"
	"net"
	"os"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

const (
	defaultDatabase = "testdb"
	defaultUser     = "test"
	defaultPassword = "test"
)

// Server wraps an embedded-postgres instance bound to a dynamic local port.
type Server struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string

	db         *embeddedpostgres.EmbeddedPostgres
	runtimeDir string
}

// Start launches an in-process PostgreSQL server on 127.0.0.1 with a free port.
func Start() (*Server, error) {
	port, err := freeTCPPort()
	if err != nil {
		return nil, err
	}

	runtimeDir, err := os.MkdirTemp("", "data-nexus-pg-*")
	if err != nil {
		return nil, err
	}

	cfg := embeddedpostgres.DefaultConfig().
		Username(defaultUser).
		Password(defaultPassword).
		Database(defaultDatabase).
		Port(uint32(port)).
		RuntimePath(runtimeDir).
		Logger(io.Discard)

	db := embeddedpostgres.NewDatabase(cfg)
	if err := db.Start(); err != nil {
		_ = os.RemoveAll(runtimeDir)
		return nil, fmt.Errorf("postgres test server: %w", err)
	}

	return &Server{
		Host:       "127.0.0.1",
		Port:       port,
		Database:   defaultDatabase,
		User:       defaultUser,
		Password:   defaultPassword,
		db:         db,
		runtimeDir: runtimeDir,
	}, nil
}

// Close stops the embedded PostgreSQL process and removes its runtime directory.
func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	var stopErr error
	if s.db != nil {
		stopErr = s.db.Stop()
	}
	if s.runtimeDir != "" {
		if err := os.RemoveAll(s.runtimeDir); err != nil && stopErr == nil {
			stopErr = err
		}
	}
	return stopErr
}

func freeTCPPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		return 0, err
	}
	return port, nil
}
