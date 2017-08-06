package sql

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type scoresRepo struct {
	db *db.DB

	log log.Log
}

func (r scoresRepo) Get(article content.Article) (content.Scores, error) {
	if err := article.Validate(); err != nil {
		return content.Scores{}, errors.WithMessage(err, "validating article")
	}

	r.log.Infof("Getting scores for article %s", article)

	var scores content.Scores
	if err := r.db.Get(&scores, r.db.SQL().Article.GetScores, article.ID); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Scores{}, errors.Wrapf(err, "getting scores for article %s", article)
	}

	return scores, nil
}

func (r scoresRepo) Update(scores content.Scores) error {
	if err := scores.Validate(); err != nil {
		return errors.WithMessage(err, "validating scores")
	}

	r.log.Infof("Updating scores %s", scores)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.Preparex(s.Article.UpdateScores)
	if err != nil {
		return errors.Wrap(err, "preparing scores update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(scores.Score, scores.Score1, scores.Score2, scores.Score3, scores.Score4, scores.Score5, scores.ArticleID)
	if err != nil {
		return errors.Wrap(err, "executimg scores update stmt")
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "committing transaction")
		}

		return nil
	}

	stmt, err = tx.Preparex(s.Article.CreateScores)
	if err != nil {
		return errors.Wrap(err, "preparing scores create stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(scores.ArticleID, scores.Score, scores.Score1, scores.Score2, scores.Score3, scores.Score4, scores.Score5)
	if err != nil {
		return errors.Wrap(err, "executimg scores create stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}
