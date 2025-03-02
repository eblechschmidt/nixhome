package server

import (
	"bytes"
	"errors"
	"html/template"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/eblechschmidt/nixhome/internal/cfg"
	"github.com/eblechschmidt/nixhome/internal/icon"
	"github.com/eblechschmidt/nixhome/web"
	"github.com/rs/zerolog/log"
)

type Server struct {
	http.Server
	cfg   *cfg.Config
	tmpl  *template.Template
	icons map[string]*icon.Icon
}

func (s *Server) serveIcon(w http.ResponseWriter, r *http.Request) {
	icon := strings.TrimPrefix(r.URL.Path, "/icons/")

	i, ok := s.icons[icon]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	colored := r.URL.Query().Has("colored")
	log.Debug().Str("icon", icon).Bool("colored", colored).Msg("Requesting image")

	w.Header().Set("Content-Type", i.Mime)
	var err error
	if colored {
		err = i.WriteColored(w)
	} else {
		err = i.Write(w)
	}
	if err != nil {
		log.Error().Err(err).Msg("Could not write colored image to response writer")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (s *Server) icon(icon string) template.HTML {
	log.Debug().Str("icon", icon).Msg("Render icon")
	b := bytes.Buffer{}
	i, ok := s.icons[icon]
	if !ok {
		return ""
	}

	err := i.Write(&b)
	if err != nil {
		return ""
	}

	return template.HTML(b.String())
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
		log.Debug().Str("mime-type", mt).Str("path", r.URL.Path).Str("filename", filename).Msg("Serving file")
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
	c, err := cfg.FromFile(cfgFile)
	if err != nil {
		return nil, err
	}

	log.Debug().Any("Colors", c.Colors).Msg("Colors")

	icons := make(map[string]*icon.Icon)
	for _, group := range c.Apps {
		for _, app := range group {
			if _, ok := icons[app.Icon]; ok || app.Icon == "" {
				continue
			}

			i, err := icon.New(app.Icon, dataDir)
			if err != nil {
				log.Error().Err(err).Msg("Could not cache icon")
			}
			icons[app.Icon] = i
		}
	}

	s := Server{
		cfg:   c,
		icons: icons,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/style.css", s.serveFile("style.css"))
	mux.HandleFunc("/icons/", s.serveIcon)

	mux.HandleFunc("/", s.serveFile("index.tmpl"))

	s.Handler = mux
	s.Addr = addr
	s.tmpl, err = template.New("index.tmpl").Funcs(
		template.FuncMap{
			"icon": s.icon,
		},
	).ParseFS(web.FS, "*")
	if err != nil {
		return nil, err
	}

	return &s, nil
}
