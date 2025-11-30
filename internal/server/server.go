package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Server struct {
	Port       int
	Folder     string
	httpServer *http.Server
	logFile    *os.File
	errorFile  *os.File
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func New(port int, folder string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Port:   port,
		Folder: folder,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Setup logging
	logsDir := filepath.Join(s.Folder, "logs")
	os.MkdirAll(logsDir, 0755)

	var err error
	s.logFile, err = os.OpenFile(
		filepath.Join(logsDir, "access.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}

	s.errorFile, err = os.OpenFile(
		filepath.Join(logsDir, "error.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}

	// Create file server with logging
	fs := http.FileServer(http.Dir(s.Folder))
	handler := s.loggingMiddleware(fs)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		s.logError(fmt.Sprintf("Server started on port %d", s.Port))
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			s.logError(fmt.Sprintf("Server error: %v", err))
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logError(fmt.Sprintf("Server shutdown error: %v", err))
		}
		s.logError("Server stopped")
	}

	if s.logFile != nil {
		s.logFile.Close()
	}
	if s.errorFile != nil {
		s.errorFile.Close()
	}

	s.cancel()
	return nil
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logMsg := fmt.Sprintf("[%s] %s %s - %v\n",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			duration,
		)

		if s.logFile != nil {
			s.logFile.WriteString(logMsg)
		}
	})
}

func (s *Server) logError(msg string) {
	logMsg := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	if s.errorFile != nil {
		s.errorFile.WriteString(logMsg)
	}
	log.Print(msg)
}
