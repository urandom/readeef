package base

import (
	"testing"

	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestArticleScores(t *testing.T) {
	a := ArticleScores{}
	a.data.Score1 = 1
	a.data.Score2 = 2
	a.data.Score3 = 3
	a.data.Score4 = 4
	a.data.Score5 = 5

	d := a.Data()

	tests.CheckInt64(t, 2, a.data.Score2)
	d.Score2 = 10
	tests.CheckInt64(t, 2, a.data.Score2)

	d = a.Data(d)

	tests.CheckInt64(t, 10, d.Score2)

	tests.CheckBool(t, false, a.Validate() == nil)

	d.ArticleId = data.ArticleId(1)
	a.Data(d)

	tests.CheckBool(t, true, a.Validate() == nil)

	tests.CheckString(t, "Scores for article '1'", a.String())

}
