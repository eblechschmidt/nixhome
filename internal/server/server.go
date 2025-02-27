package server

import (
	"errors"
	"html/template"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/eblechschmidt/nixhome/internal/cfg"
	"github.com/eblechschmidt/nixhome/web"
	"github.com/rs/zerolog/log"
)

type Server struct {
	http.Server
	c *cfg.Config
	t *template.Template
}

func (s *Server) serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := web.FS.Open(filename)
		if err != nil {
			log.Error().Err(err).Msg("Error serving file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ext := filepath.Ext(filename)
		log.Debug().Str("ext", ext).Msg("Serving file")
		switch ext {
		case ".css":
			t := mime.TypeByExtension(ext)
			log.Debug().Str("mime-type", t).Msg("Serving file")
			w.Header().Set("Content-Type", t)
		}

		_, err = io.Copy(w, f)
		if err != nil {
			log.Error().Err(err).Msg("Error serving file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debug().Str("path", r.URL.Path).Str("filename", filename).Msg("Serving file")
	}
}

func (s *Server) Serve() error {
	err := s.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err

}

func New(cfgFile, addr string) (*Server, error) {
	c, err := cfg.FromFile(cfgFile)
	if err != nil {
		return nil, err
	}

	s := Server{
		c: c,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/style.css", s.serveFile("style.css"))
	mux.HandleFunc("/", s.serveFile("index.tmpl"))

	s.Handler = mux
	s.Addr = addr
	s.t, err = template.New("index.tmpl").ParseFS(web.FS, "*")
	if err != nil {
		return nil, err
	}

	return &s, nil
}
