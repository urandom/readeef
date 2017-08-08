package readeef

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/pkg/errors"

	"golang.org/x/text/language"
)

var (
	languagesLoaded = false
)

// GetLanguages returns a slice of supported languages
func GetLanguages(fs http.FileSystem) ([]language.Tag, error) {
	locale, err := fs.Open("locale")
	if cause := errors.Cause(err); cause == os.ErrNotExist {
		return []language.Tag{}, nil
	} else if err != nil {
		return nil, errors.WithMessage(err, "opening locale dir")
	} else {
		if langFiles, err := locale.Readdir(-1); err == nil {
			languages := make([]language.Tag, 0, len(langFiles))
			for _, f := range langFiles {
				name := f.Name()
				suffix := ".all.json"
				if strings.HasSuffix(name, suffix) {
					if !languagesLoaded {
						var b []byte
						f, err := fs.Open(path.Join("locale", name))
						if err == nil {
							b, err = ioutil.ReadAll(f)
						}
						if err != nil {
							return nil, errors.Wrapf(err, "reading /locale/%s", name)
						}

						if err = i18n.ParseTranslationFileBytes(name, b); err != nil {
							return nil, errors.Wrapf(err, "parsing language file %s", name)
						}
					}

					tag := language.Make(name[:len(name)-len(suffix)])
					if !tag.IsRoot() {
						languages = append(languages, tag)
					}
				}
			}

			languagesLoaded = true
			return languages, nil
		}

		return nil, errors.WithMessage(err, "reading locale dir")
	}
}
