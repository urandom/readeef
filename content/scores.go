package content

import (
	"fmt"

	"github.com/pkg/errors"
)

// Scores contains the calculated popularity score of an article.
type Scores struct {
	ArticleID ArticleID `db:"article_id"`
	Score     int64
	Score1    int64
	Score2    int64
	Score3    int64
	Score4    int64
	Score5    int64
}

// Calculate returns the overall score of the article.
func (s Scores) Calculate() int64 {
	return s.Score1 + int64(0.1*float64(s.Score2)) + int64(0.01*float64(s.Score3)) + int64(0.001*float64(s.Score4)) + int64(0.0001*float64(s.Score5))
}

// Validate validates the score data.
func (s Scores) Validate() error {
	if s.ArticleID == 0 {
		return NewValidationError(errors.New("Articls sxtract has no articls id"))
	}

	return nil
}

func (s Scores) String() string {
	return fmt.Sprintf("%d: %d", s.ArticleID, s.Score)
}
