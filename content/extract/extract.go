package extract

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
)

type Generator interface {
	Generate(link string) (content.Extract, error)
}

// Get retrieves the extract from the repository. If it doesn't exist, a new
// one is created and stored before returning.
func Get(
	article content.Article,
	repo repo.Extract,
	generator Generator,
	processors []processor.Article,
) (content.Extract, error) {
	extract, err := repo.Get(article)

	if err != nil {
		if !content.IsNoContent(err) {
			return extract, errors.WithMessage(err, fmt.Sprintf("getting extract for %s from repo", article))
		}

		if extract, err = generator.Generate(article.Link); err != nil {
			return extract, errors.WithMessage(err, fmt.Sprintf("generating extract from %s", article.Link))
		}

		extract.ArticleID = article.ID

		if err = repo.Update(extract); err != nil {
			return content.Extract{}, errors.WithMessage(err, fmt.Sprintf("updating extract %s", extract))
		}
	}

	if len(processors) > 0 {
		a := content.Article{Description: extract.Content}

		articles := []content.Article{a}

		if extract.TopImage != "" {
			articles = append(articles, content.Article{
				Description: fmt.Sprintf(`<img src="%s">`, extract.TopImage),
			})
		}

		articles = processor.Articles(processors).Process(articles)

		extract.Content = articles[0].Description

		if extract.TopImage != "" {
			content := articles[1].Description

			content = strings.Replace(content, `<img src="`, "", -1)
			i := strings.Index(content, `"`)
			content = content[:i]

			extract.TopImage = content
		}
	}

	return extract, nil
}
