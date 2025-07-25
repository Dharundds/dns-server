package server

import (
	"dns-server/internal/api"
	"dns-server/internal/constants"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Server struct {
	addr        string
	port        int
	annInterval int
	name        string
	id          uuid.UUID
}

type ServerOption func(*Server)

func NewServer(options ...ServerOption) *Server {
	s := &Server{
		id:          uuid.New(),
		name:        "DP Home Server",
		addr:        "",
		port:        8080,
		annInterval: 30,
	}

	for _, option := range options {
		option(s)
	}

	return s
}

func WithName(name string) ServerOption {
	return func(s *Server) {
		s.name = name
	}
}

func WithAnnouncementInterval(interval int) ServerOption {
	return func(s *Server) {
		s.annInterval = interval
	}
}

func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) GetAddr() string {
	return s.addr
}

func (s *Server) GetPort() int {
	return s.port
}

func (s *Server) GetName() string {
	return s.name
}

func (s *Server) GetUUID() string {
	return s.id.String()
}

func StartFrontend() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Construct the file path based on the request URL
		path := filepath.Join(constants.BuildPath, r.URL.Path)
		path = filepath.Clean(path)
		// Check if the requested file exists
		_, err := os.Stat(path)
		if os.IsNotExist(err) || r.URL.Path == "/" {
			// If the file does not exist or it's the root path, serve index.html
			http.ServeFile(w, r, filepath.Join(constants.BuildPath, "index.html"))
		} else {
			// Otherwise, serve the static file
			http.FileServer(http.Dir(constants.BuildPath)).ServeHTTP(w, r)
		}
	})
	log.Info().Msgf("Frontend server starting on 3000")
	return http.ListenAndServe(":3000", nil)

}

func (s *Server) StartBackend() error {
	g := gin.Default()
	// g.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	// c.Header("Access-Control-Allow-Headers", "Content-Type, SOAPAction")
	// c.Header("Access-Control-Max-Age", "86400")

	// Configure CORS
	g.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},                                                           // Allowed origins
		AllowMethods:  []string{"GET", "POST", "DELETE", "OPTIONS"},                            // Allowed HTTP methods
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},                     // Allowed headers
		ExposeHeaders: []string{"Content-Length", "Access-Control-Allow-Origin", "SOAPAction"}, // Headers exposed to the browser
	}))

	api.HandleFuncs(g)
	addr := fmt.Sprintf("%s:%d", s.addr, s.port)
	return g.Run(addr)
}
