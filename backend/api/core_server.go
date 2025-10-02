package api

import (
	"brainbook-api/api/websocket"
	"brainbook-api/internal/database"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

/*
Defines a config struct to hold all Application configuration settings.
values are set from env variables when the Application starts.
*/
type Config struct {
	BaseURL  string
	HttpPort int
	DB       struct {
		DSN         string
		Automigrate bool
	}
	// JWT struct {
	// 	SecretKey string
	// }
}

// Application holds the dependencies for HTTP handlers, helpers,
// and middleware.
type Application struct {
	Config    Config
	DB        *database.DB
	Logger    *slog.Logger
	WG        sync.WaitGroup
	WSManager *websocket.WebsocketManager
}

const (
	defaultIdleTimeout    = time.Minute
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 10 * time.Second
	defaultShutdownPeriod = 30 * time.Second
)

func (app *Application) ServeHTTP() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.HttpPort),
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(app.Logger.Handler(), slog.LevelWarn),
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	shutdownErrorChan := make(chan error)

	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownPeriod)
		defer cancel()

		shutdownErrorChan <- srv.Shutdown(ctx)
	}()

	app.Logger.Info("starting server", slog.Group("server", "addr", srv.Addr))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErrorChan
	if err != nil {
		return err
	}

	app.Logger.Info("stopped server", slog.Group("server", "addr", srv.Addr))

	app.WG.Wait()
	return nil
}

// neuteredFileHandler creates a custom file handler that returns JSON error responses
// instead of HTML when files are not found or directories are accessed inappropriately.
func (app *Application) neuteredFileHandler(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Cleans the path to prevent directory traversal attacks.
		cleanPath := filepath.Clean(r.URL.Path)
		if strings.Contains(cleanPath, "..") {
			app.notFound(w, r)
			return
		}

		// Builds the full file path.
		fullPath := filepath.Join(dir, cleanPath)

		// Checks if the file exists and retrieves its info.
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			// Returns a JSON 404 response if the file doesn't exist.
			app.notFound(w, r)
			return
		}

		// Serves index.html if it exists and the URL requested corresponds to a directory.
		if fileInfo.IsDir() {
			indexPath := filepath.Join(fullPath, "index.html")
			if _, err := os.Stat(indexPath); err != nil {
				// Sends a JSON 404 response when no index.html is found in the directory.
				app.notFound(w, r)
				return
			}
			// Serves the index.html file.
			fullPath = indexPath
		}

		ext := filepath.Ext(fullPath)
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".html":
			w.Header().Set("Content-Type", "text/html")
		}

		http.ServeFile(w, r, fullPath)
	}
}

// neuteredFileSystem is a custom type which embeds the standard http.FileSystem.
// type neuteredFileSystem struct {
// 	fs http.FileSystem
// }

// Open checks the requested file path and determines whether it is a directory or not.
// If it is a directory, it attempts to open an index.html file in it.
// func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
// 	f, err := nfs.fs.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	s, err := f.Stat()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if s.IsDir() {
// 		index := filepath.Join(path, "index.html")
// 		if _, err := nfs.fs.Open(index); err != nil {
// 			// Closes the original file to avoid a file descriptor leak
// 			closeErr := f.Close()
// 			if closeErr != nil {
// 				return nil, closeErr
// 			}

// 			return nil, err
// 		}
// 	}

// 	return f, nil
// }
