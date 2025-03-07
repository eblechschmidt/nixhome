package icon

import (
	"crypto/sha256"
	"fmt"
	"image/color"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/crazy3lf/colorconv"
	"github.com/rs/zerolog/log"
)

func New(icon string, dataDir string, col string) (string, error) {
	if icon == "" {
		log.Info().Msg("No icon specified")
		return "", nil
	}
	dir := filepath.Join(dataDir, "icons")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}

	fname, err := cacheFile(icon, dir)
	if err != nil {
		return "", err
	}

	if fname == "" {
		fname, err = download(icon, dir)
		if err != nil {
			return "", err
		}
	}

	f, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return fixViewBox(string(b)), nil
}

func download(icon, dir string) (string, error) {
	log.Info().Str("icon", icon).Str("dir", dir).Msg("Downloading icon")
	ispath := strings.Count(icon, "/") > 0
	var url string
	switch {
	case strings.HasPrefix(icon, "http"):
		url = icon
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

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("icon not found")
	}
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
		log.Info().Str("icon", icon).Str("url", url).Strs("content-type", ct).Strs("ext", e).Msg("Get header")
	}
	log.Info().Str("icon", icon).Str("status", resp.Status).Str("url", url).Strs("content-type", ct).Msg("Get header")

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

	return fname, nil
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
			log.Info().Str("icon", icon).Str("filename", f).Msg("Cashed icon found")
			return f, nil
		}
	}
	log.Info().Str("icon", icon).Str("dir", dir).Msg("Icon not found in cash")
	return "", nil
}

func base(fname string) string {
	return strings.TrimSuffix(filepath.Base(fname), filepath.Ext(fname))
}

func hex2col(hex string) color.Color {
	if len(hex) == 4 {
		hex = fmt.Sprintf("#%c%c%c%c%c%c", hex[1], hex[1], hex[2], hex[2], hex[3], hex[3])
	}
	col, err := colorconv.HexToColor(hex)
	if err != nil {
		fmt.Println(err)
		panic("this should not happen")
	}
	return col
}

func col2hex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func Colorize(data, color string) string {
	hcol, scol, lcol := colorconv.ColorToHSL(hex2col(color))
	// Regular expression to match both short and long hex color codes

	re := regexp.MustCompile(`#[0-9a-fA-F]{3,6}\b`)

	// Find all matches

	matches := re.FindAllString(data, -1)

	lmin := 1.0

	for _, m := range matches {
		_, _, l := colorconv.ColorToHSL(hex2col(m))

		if l < lmin {
			lmin = l
		}
	}

	for _, m := range matches {
		_, _, l := colorconv.ColorToHSL(hex2col(m))
		lnew := 1 - ((1 - l) / (1 - lmin) * (1 - lcol))
		// fmt.Println(l, lmin, lcol, lnew)
		newcol, err := colorconv.HSLToColor(hcol, scol, lnew)
		if err != nil {
			fmt.Println(err)
			panic("this should not have happened")
		}
		newhex := col2hex(newcol)

		data = strings.ReplaceAll(data, m, newhex)

	}

	// svg that do not have color information will be colored directly with a
	// style attribute to the paths
	if len(matches) == 0 {
		data = strings.ReplaceAll(data, "<path ", fmt.Sprintf("<path style=\"fill:%s\" ", color))
	}

	data = strings.Trim(data, "\n\r\t")
	return data
}

func fixViewBox(data string) string {
	re := regexp.MustCompile(`<svg[^>]*\bviewBox="([^"]+)"`)
	match := re.FindStringSubmatch(data)
	if len(match) > 0 {
		return data
	}
	fmt.Println(match)

	re1 := regexp.MustCompile(`<svg[^>]*\bwidth="([^"]+)"`)
	match = re1.FindStringSubmatch(data)
	if len(match) < 2 {
		return data
	}
	width := strings.TrimSuffix(match[1], "px")
	fmt.Println(width)

	re = regexp.MustCompile(`<svg[^>]*\bwidth="([^"]+)"`)
	match = re.FindStringSubmatch(data)
	if len(match) < 2 {
		return data
	}
	height := strings.TrimSuffix(match[1], "px")

	return strings.ReplaceAll(data, "<svg ", fmt.Sprintf("<svg viewBox=\"0 0 %s %s\" ", width, height))
}
