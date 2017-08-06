package content

import (
	"errors"
	"fmt"
)

type Thumbnail struct {
	ArticleID ArticleID `db:"article_id"`
	Thumbnail string
	Link      string
	Processed bool
}

func (t Thumbnail) Validate() error {
	if t.ArticleID == 0 {
		return NewValidationError(errors.New("Article thumbnail has no article id"))
	}

	return nil
}

func (t Thumbnail) String() string {
	return fmt.Sprintf("%d: %s", t.ArticleID, t.Link)
}
