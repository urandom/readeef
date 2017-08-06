package content

import (
	"errors"
	"fmt"
)

type Extract struct {
	ArticleID ArticleID `db:"article_id"`
	Title     string
	Content   string
	TopImage  string `db:"top_image"`
	Language  string
}

func (e Extract) Validate() error {
	if e.ArticleID == 0 {
		return NewValidationError(errors.New("Article extract has no article id"))
	}

	return nil
}

func (e Extract) String() string {
	return fmt.Sprintf("%d: %s", e.ArticleID, e.Title)
}
