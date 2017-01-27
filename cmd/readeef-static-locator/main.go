package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	css    = "/css/"
	dist   = "/dist/"
	js     = "/js/"
	images = "/images/"
)

var (
	output      string
	templateDir string
	staticDir   string
	localeDir   string
	singleLine  bool

	sqstring      = regexp.MustCompile(`'([^\\']*(?:(?:\\'|\\)[^\\']*)*)'`)
	dqstring      = regexp.MustCompile(`"([^\\"]*(?:(?:\\"|\\)[^\\"]*)*)"`)
	importScripts = regexp.MustCompile(`importScripts\(([^)]+)\)`)
	cssURL        = regexp.MustCompile(`url\(([^)]+)\)`)
	worker        = regexp.MustCompile(`new Worker\(([^)]+)\)`)

	seen   = make(map[string]bool)
	parsed = make(map[string]bool)
)

func main() {
	flag.Parse()

	var out *os.File

	if output == "-" {
		out = os.Stdout
	} else {
		var err error

		if out, err = os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening '%s' for writing: %v\n", output, err)
			os.Exit(0)
		}

		defer out.Close()
	}

	templates := []string{}
	filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			templates = append(templates, path)
		}

		return nil
	})

	static := []string{}
	for _, p := range templates {
		if !strings.HasSuffix(p, ".tmpl") && !strings.HasSuffix(p, ".html") {
			fmt.Fprintf(os.Stderr, "Skipping file '%s'\n", p)
		}

		s, err := parseHTMLFile(p)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(0)
		}

		static = append(static, s...)
	}

	locales := []string{}
	filepath.Walk(localeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s: %v\n", path, err)
			return nil
		}
		if !info.IsDir() {
			locales = append(locales, path)
		}

		return nil
	})

	sort.Strings(templates)
	sort.Strings(static)
	sort.Strings(locales)

	buf := new(bytes.Buffer)
	delim := "\n"
	if singleLine {
		delim = " "
	}

	for _, t := range templates {
		buf.WriteString(t + delim)
	}

	buf.WriteString(delim)
	for _, s := range static {
		buf.WriteString(s + delim)
	}

	buf.WriteString(delim)
	for _, s := range locales {
		buf.WriteString(s + delim)
	}

	buf.WriteTo(out)
}

func parseHTMLFile(path string) (static []string, err error) {
	if parsed[path] {
		return
	}
	parsed[path] = true

	var f *os.File
	if f, err = os.Open(path); err != nil {
		err = errors.New(fmt.Sprintf("Error opening '%s': %v\n", path, err))
		return
	}
	defer f.Close()

	var d *goquery.Document
	if d, err = goquery.NewDocumentFromReader(f); err != nil {
		err = errors.New(fmt.Sprintf("Error parsing '%s': %v\n", path, err))
		return
	}

	d.Find("[src], [href]").Each(func(i int, s *goquery.Selection) {
		var val string
		var ok bool

		if val, ok = s.Attr("src"); !ok {
			if val, ok = s.Attr("href"); !ok {
				return
			}
		}

		// Template expansion or other undesirables
		if val[0] == '{' || val[0] == '[' || val[0] == '#' || strings.Index(val, "://") != -1 {
			return
		}

		if val[0] != '/' {
			dir, _ := filepath.Split(path)
			if strings.HasPrefix(dir, staticDir) {
				dir = dir[len(staticDir):]
			}

			val = filepath.Join(dir, val)
		}

		var p string
		if strings.HasPrefix(val, css) ||
			strings.HasPrefix(val, dist) ||
			strings.HasPrefix(val, images) ||
			strings.HasPrefix(val, js) {

			p = filepath.Join(staticDir, val)
		} else {
			return
		}

		if !seen[p] {
			static = append(static, p)
			seen[p] = true
		}
	})

	d.Find("style").Each(func(i int, s *goquery.Selection) {
		var st []string

		st, err = parseCSSContent(path, s.Text())
		if err == nil {
			static = append(static, st...)
		}
	})

	if err != nil {
		return
	}

	d.Find("script").Each(func(i int, s *goquery.Selection) {
		if _, ok := s.Attr("src"); ok {
			return
		}

		var st []string

		st, err = parseJSContent(path, s.Text())
		if err == nil {
			static = append(static, st...)
		}
	})

	if err != nil {
		return
	}

	for _, p := range static {
		var s []string

		if strings.HasSuffix(p, ".html") {
			s, err = parseHTMLFile(p)
		} else if strings.HasSuffix(p, ".js") {
			s, err = parseJSFile(p)
		} else if strings.HasSuffix(p, ".css") {
			s, err = parseCSSFile(p)
		}

		if err != nil {
			err = fmt.Errorf("Error while parsing '%s': %v\n", path, err)
			return
		}

		static = append(static, s...)
	}

	return
}

func parseJSFile(path string) (static []string, err error) {
	if parsed[path] {
		return
	}
	parsed[path] = true

	var b []byte

	b, err = ioutil.ReadFile(path)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error opening '%s': %v\n", path, err))
		return
	}

	return parseJSContent(path, string(b))
}

func parseCSSFile(path string) (static []string, err error) {
	if parsed[path] {
		return
	}
	parsed[path] = true

	var b []byte

	b, err = ioutil.ReadFile(path)
	if err != nil {
		err = errors.New(fmt.Sprintf("Error opening '%s': %v\n", path, err))
		return
	}

	return parseCSSContent(path, string(b))
}

func parseJSContent(path, content string) (static []string, err error) {
	matches := importScripts.FindAllStringSubmatch(content, -1)
	matches = append(matches, worker.FindAllStringSubmatch(content, -1)...)
	for _, match := range matches {
		if len(match) > 1 {
			args := match[1]

			scripts := sqstring.FindAllString(args, -1)
			scripts = append(scripts, dqstring.FindAllString(args, -1)...)

			for _, val := range scripts {
				val = val[1 : len(val)-1]
				if val[0] != '/' {
					dir, _ := filepath.Split(path)
					if strings.HasPrefix(dir, staticDir) {
						dir = dir[len(staticDir):]
					}

					val = filepath.Join(dir, val)
				}

				var p string
				if strings.HasPrefix(val, css) ||
					strings.HasPrefix(val, dist) ||
					strings.HasPrefix(val, images) ||
					strings.HasPrefix(val, js) {

					p = filepath.Join(staticDir, val)
				} else {
					continue
				}

				if !seen[p] {
					static = append(static, p)
					seen[p] = true
				}
			}
		}
	}

	for _, p := range static {
		var s []string

		if strings.HasSuffix(p, ".js") {
			s, err = parseJSFile(p)
		}

		if err != nil {
			err = fmt.Errorf("Error while parsing '%s': %v\n", path, err)
			return
		}

		static = append(static, s...)
	}

	return
}

func parseCSSContent(path, content string) (static []string, err error) {
	matches := cssURL.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			args := match[1]

			urls := sqstring.FindAllString(args, -1)
			urls = append(urls, dqstring.FindAllString(args, -1)...)

			for _, val := range urls {
				val = val[1 : len(val)-1]
				if val[0] != '/' {
					dir, _ := filepath.Split(path)
					if strings.HasPrefix(dir, staticDir) {
						dir = dir[len(staticDir):]
					}

					val = filepath.Join(dir, val)
				}

				var p string
				if strings.HasPrefix(val, css) ||
					strings.HasPrefix(val, dist) ||
					strings.HasPrefix(val, images) ||
					strings.HasPrefix(val, js) {

					p = filepath.Join(staticDir, val)
				} else {
					continue
				}

				if !seen[p] {
					static = append(static, p)
					seen[p] = true
				}
			}
		}
	}

	for _, p := range static {
		var s []string

		if strings.HasSuffix(p, ".css") {
			s, err = parseCSSFile(p)
		}

		if err != nil {
			err = fmt.Errorf("Error while parsing '%s': %v\n", path, err)
			return
		}

		static = append(static, s...)
	}

	return
}

func init() {
	flag.StringVar(&output, "output", "-", "the output file")
	flag.BoolVar(&singleLine, "single-line", false, "list all files in a single line")
	flag.StringVar(&templateDir, "template-dir", "templates", "the templates directory")
	flag.StringVar(&staticDir, "static-dir", "static", "the static directory")
	flag.StringVar(&localeDir, "locale-dir", "locale", "the locale directory")
}
