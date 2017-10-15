package thumbnail

import (
	"fmt"
	_ "image/png"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/extract"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type ext struct {
	repo        repo.Thumbnail
	extractRepo repo.Extract
	generator   extract.Generator
	processors  []processor.Article
	log         log.Log
}

func FromExtract(
	repo repo.Thumbnail,
	extractRepo repo.Extract,
	g extract.Generator,
	processors []processor.Article,
	log log.Log,
) (Generator, error) {
	if g == nil {
		return nil, errors.New("no extract.Generator")
	}

	processors = filterProcessors(processors)

	return ext{repo: repo, extractRepo: extractRepo, generator: g, processors: processors, log: log}, nil
}

func (t ext) Generate(a content.Article) error {
	thumbnail := content.Thumbnail{ArticleID: a.ID, Processed: true}

	t.log.Debugf("Generating thumbnail for article %s from extract", a)

	thumbnail.Thumbnail, thumbnail.Link =
		generateThumbnailFromDescription(strings.NewReader(a.Description))

	if thumbnail.Link == "" {
		t.log.Debugf("%s description doesn't contain suitable link, getting extract\n", a)

		extract, err := extract.Get(a, t.extractRepo, t.generator, t.processors)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("getting article extract for %s", a))
		}

		if extract.TopImage == "" {
			t.log.Debugf("Extract for %s doesn't contain a top image", a)
		} else {
			t.log.Debugf("Generating thumbnail from top image %s of %s\n", extract.TopImage, a)
			thumbnail.Thumbnail = generateThumbnailFromImageLink(extract.TopImage)
			thumbnail.Link = extract.TopImage
		}
	}

	if err := t.repo.Update(thumbnail); err != nil {
		return errors.WithMessage(err, "saving thumbnail to repo")
	}

	return nil
}

func filterProcessors(input []processor.Article) []processor.Article {
	processors := make([]processor.Article, 0, len(input))

	for i := range input {
		if _, ok := input[i].(processor.ProxyHTTP); ok {
			continue
		}

		processors = append(processors, input[i])
	}

	return processors
}
