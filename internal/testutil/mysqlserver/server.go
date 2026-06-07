// Package mysqlserver starts an in-process MySQL-compatible server for integration tests.
package mysqlserver

import (
	"fmt"
	"io"
	"net"
	"sync"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/sirupsen/logrus"
)

const defaultDatabase = "testdb"

// Server wraps a go-mysql-server instance bound to a dynamic local port.
type Server struct {
	Host     string
	Port     int
	Database string

	srv *server.Server
	wg  sync.WaitGroup
}

// Start launches an in-process MySQL-compatible server on 127.0.0.1 with a free port.
func Start() (*Server, error) {
	logrus.SetOutput(io.Discard)

	port, err := freeTCPPort()
	if err != nil {
		return nil, err
	}

	db := memory.NewDatabase(defaultDatabase)
	pro := memory.NewDBProvider(db)
	engine := sqle.NewDefault(pro)

	cfg := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("127.0.0.1:%d", port),
	}
	srv, err := server.NewServer(cfg, engine, sql.NewContext, memory.NewSessionBuilder(pro), nil)
	if err != nil {
		return nil, fmt.Errorf("mysql test server: %w", err)
	}

	s := &Server{
		Host:     "127.0.0.1",
		Port:     port,
		Database: defaultDatabase,
		srv:      srv,
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		_ = srv.Start()
	}()
	return s, nil
}

// Close stops the server listener.
func (s *Server) Close() error {
	if s == nil || s.srv == nil {
		return nil
	}
	err := s.srv.Close()
	s.wg.Wait()
	return err
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
