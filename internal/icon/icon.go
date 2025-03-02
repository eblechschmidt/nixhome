package icon

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/eblechschmidt/nixhome/internal/cfg"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

type Icon struct {
	data   *html.Node
	Mime   string
	colors *cfg.Colors
}

func (i *Icon) Write(w io.Writer) error {
	// _, err := w.Write(i.data)
	err := render(w, i.data)
	if err != nil {
		return err
	}
	return nil
}

func (i *Icon) WriteColored(w io.Writer) error {
	// if !bytes.Contains(i.data, []byte("fill:")) {
	// 	ind := bytes.Index(i.data, []byte(">"))
	// 	_, err := w.Write(i.data[:ind])
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_, err = w.Write([]byte(" style=\"fill: var(--color-text-pri);\""))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_, err = w.Write(i.data[ind:])
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }

	err := render(w, i.data)
	// _, err := w.Write(i.data)
	if err != nil {
		return err
	}
	return nil

}

func New(icon string, dataDir string) (*Icon, error) {
	if icon == "" {
		log.Debug().Msg("No icon specified")
		return nil, nil
	}
	dir := filepath.Join(dataDir, "icons")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	fname, err := cacheFile(icon, dir)
	if err != nil {
		return nil, err
	}

	if fname == "" {
		fname, err = download(icon, dir)
		if err != nil {
			return nil, err
		}
	}

	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h, err := html.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("could not parse svg icon: %w", err)
	}
	removeSize(h)

	m := mime.TypeByExtension(filepath.Ext(fname))

	return &Icon{data: h, Mime: m}, nil
}

func download(icon, dir string) (string, error) {
	log.Debug().Str("icon", icon).Str("dir", dir).Msg("Downloading icon")
	ispath := strings.Count(icon, "/") > 0
	var url string
	switch {
	case !ispath:
		url = fmt.Sprintf("https://raw.githubusercontent.com/simple-icons/simple-icons/refs/heads/develop/icons/%s.svg", icon)
	case ispath:
		url = fmt.Sprintf("https://www.svgrepo.com/download/%s.svg", icon)
	default:
		return "", fmt.Errorf("icon '%s' not valid", icon)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ct, ok := resp.Header["Content-Type"]
	ext := ".txt"
	if ok && len(ct) == 1 {
		e, err := mime.ExtensionsByType(ct[0])
		if err != nil {
			return "", err
		}
		if len(e) > 0 {
			ext = e[0]
		}
		log.Debug().Str("icon", icon).Str("url", url).Strs("content-type", ct).Strs("ext", e).Msg("Get header")
	}

	fname := filepath.Join(dir, hash(icon)+ext)
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}

	return "", nil
}

func hash(s string) string {
	h := sha256.New()

	h.Write([]byte(s))
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

func cacheFile(icon, dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	h := hash(icon)
	for _, file := range files {
		if base(file.Name()) == h {
			f := filepath.Join(dir, file.Name())
			log.Debug().Str("icon", icon).Str("dir", dir).Str("filename", f).Msg("Cashed icon found")
			return f, nil
		}
	}
	log.Debug().Str("icon", icon).Str("dir", dir).Msg("Icon not found in cash")
	return "", nil
}

func base(fname string) string {
	return strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname))
}

func removeSize(n *html.Node) {
	if n.Type == html.ElementNode {
		attr := make([]html.Attribute, 0, len(n.Attr))
		for _, a := range n.Attr {
			if (a.Key != "width") && (a.Key != "height") {
				attr = append(attr, a)
			}
		}
		n.Attr = attr
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		removeSize(c)
	}
}

func render(w io.Writer, n *html.Node) error {
	log.Debug().Str("Data", n.Data).Uint32("Type", uint32(n.Type)).Msg("Render node")
	tag := ""
	if n.Type == html.ElementNode &&
		(n.Data != "html" && n.Data != "head" && n.Data != "body") {
		_, err := io.WriteString(w, fmt.Sprintf("<%s", n.Data))
		if err != nil {
			log.Error().Err(err).Msg("Could not write start of tag")
			return err
		}
		tag = n.Data

		for _, a := range n.Attr {
			if (a.Key != "width") && (a.Key != "height") {
				_, err := io.WriteString(w, fmt.Sprintf(" %s=\"%s\"", a.Key, a.Val))
				if err != nil {
					log.Error().Err(err).Msg("Attr")
					return err
				}
			}
		}
		_, err = io.WriteString(w, ">")
		if err != nil {
			log.Error().Err(err).Msg("Could not write end of tag")
			return err
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err := render(w, c)
		if err != nil {
			log.Error().Err(err).Msg("Could not render children")
			return err
		}
	}

	if tag != "" {
		_, err := io.WriteString(w, fmt.Sprintf("</%s>", tag))
		if err != nil {
			log.Error().Err(err).Msg("Could not write closing tag")
			return err
		}
	}
	return nil
}
