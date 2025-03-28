package server

import (
	"errors"
	"html/template"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/eblechschmidt/nixhome/internal/cfg"
	"github.com/eblechschmidt/nixhome/internal/icon"
	"github.com/rs/zerolog/log"
)

type Server struct {
	http.Server
	cfg  *cfg.Config
	tmpl *template.Template
}

func (s *Server) serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := s.tmpl.Lookup(filename)
		if t == nil {
			log.Error().Str("filename", filename).Msg("Template not found")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ext := filepath.Ext(filename)
		var mt string
		switch ext {
		case ".css":
			mt = mime.TypeByExtension(ext)
		default:
			mt = mime.TypeByExtension(".html")
		}
		w.Header().Set("Content-Type", mt)

		err := t.Execute(w, s.cfg)
		if err != nil {
			log.Error().Err(err).Msg("Error serving file")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Info().Str("mime-type", mt).Str("path", r.URL.Path).Str("filename", filename).Msg("Serving file")
	}
}

func (s *Server) Serve() error {
	err := s.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err

}

func New(cfgFile, addr, dataDir string) (*Server, error) {
	log.Info().Str("cfgFile", cfgFile).Msg("Reading config")
	c, err := cfg.FromFile(cfgFile)
	if err != nil {
		return nil, err
	}

	for _, group := range c.Apps {
		for _, app := range group {
			data, err := icon.New(string(app.Icon), dataDir, c.Colors.Dark.Text)
			if err != nil {
				log.Error().Err(err).Msg("Could not cache icon")
				continue
			}
			app.Icon = template.HTML(data)
			app.ColorizedIcon = template.HTML(icon.Colorize(string(data), c.Colors.Icon))
		}
	}

	s := Server{
		cfg: c,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/style.css", s.serveFile("style.css"))

	mux.HandleFunc("/", s.serveFile("index.tmpl"))

	s.Handler = mux
	s.Addr = addr

	s.tmpl, err = template.New("index.tmpl").
		ParseFS(os.DirFS(filepath.Join(dataDir, "template")), "*")
	if err != nil {
		return nil, err
	}

	return &s, nil
}
