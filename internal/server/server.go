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
	Running    bool
}

func New(port int, folder string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Port:    port,
		Folder:  folder,
		ctx:     ctx,
		cancel:  cancel,
		Running: false,
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Running {
		return fmt.Errorf("server is already running")
	}

	// Setup logging
	logsDir := filepath.Join(s.Folder, "logs")
	os.MkdirAll(logsDir, 0755)

	var err error
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	s.logFile, err = os.OpenFile(
		filepath.Join(logsDir, fmt.Sprintf("access_%s.log", timestamp)),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}

	s.errorFile, err = os.OpenFile(
		filepath.Join(logsDir, fmt.Sprintf("error_%s.log", timestamp)),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		s.logFile.Close()
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
		s.logInfo(fmt.Sprintf("Server starting on port %d", s.Port))
		s.logInfo(fmt.Sprintf("Serving files from: %s", s.Folder))
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			s.logError(fmt.Sprintf("Server error: %v", err))
			s.Running = false
		}
	}()

	// Wait a moment to ensure server is up
	time.Sleep(100 * time.Millisecond)

	// Check if server is reachable
	if err := s.checkServer(); err != nil {
		s.Running = false
		return fmt.Errorf("server failed to start: %v", err)
	}

	s.Running = true
	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Running {
		return fmt.Errorf("server is not running")
	}

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logError(fmt.Sprintf("Server shutdown error: %v", err))
			return err
		}
		s.logInfo("Server stopped gracefully")
	}

	if s.logFile != nil {
		s.logFile.Close()
	}
	if s.errorFile != nil {
		s.errorFile.Close()
	}

	s.cancel()
	s.Running = false
	return nil
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response wrapper to capture status code
		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		logMsg := fmt.Sprintf("[%s] %s %s %d - %v\n",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			rw.status,
			duration,
		)

		if s.logFile != nil {
			s.logFile.WriteString(logMsg)
		}
	})
}

func (s *Server) logInfo(msg string) {
	logMsg := fmt.Sprintf("[%s] INFO: %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	if s.errorFile != nil {
		s.errorFile.WriteString(logMsg)
	}
	log.Print(msg)
}

func (s *Server) logError(msg string) {
	logMsg := fmt.Sprintf("[%s] ERROR: %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	if s.errorFile != nil {
		s.errorFile.WriteString(logMsg)
	}
	log.Print(msg)
}

func (s *Server) checkServer() error {
	// Try to connect to the server
	url := fmt.Sprintf("http://localhost:%d", s.Port)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Helper struct to capture response status
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
