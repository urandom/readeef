package readeef

import (
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/text/language"
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
					tag := language.Make(name[:len(name)-len(suffix)])
					if !tag.IsRoot() {
						languages = append(languages, tag)
					}
				}
			}

			return languages, nil
		}

		return nil, errors.WithMessage(err, "reading locale dir")
	}
}
