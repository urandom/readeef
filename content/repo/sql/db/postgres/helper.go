package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/content/repo/sql/db/base"
)

type Helper struct {
	*base.Helper
}

func (h Helper) InitSQL() []string {
	return initSQL
}

func (h Helper) CreateWithID(tx *sqlx.Tx, sql string, arg interface{}) (int64, error) {
	var id int64

	sql += " RETURNING id"

	stmt, err := tx.PrepareNamed(sql)
	if err != nil {
		return 0, errors.Wrap(err, "preparing create-with-id statement")
	}
	defer stmt.Close()

	err = stmt.QueryRow(arg).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "scanning created id")
	}

	return id, nil
}

func (h Helper) Upgrade(db *db.DB, old, new int) error {
	for old < new {
		var err error
		switch old {
		case 1:
			err = upgrade1to2(db)
		case 2:
			err = upgrade2to3(db)
		case 3:
			err = upgrade3to4(db)
		}

		if err != nil {
			return fmt.Errorf("Error upgrading db from %d to %d: %v\n", old, new, err)
		}
		old++
	}

	return nil
}

func (h Helper) WhereMultipleORs(column, prefix string, length int, equal bool) string {
	if length < 70 {
		return h.Helper.WhereMultipleORs(column, prefix, length, equal)
	}

	orSlice := make([]string, length)
	orSlice[0] = fmt.Sprintf("VALUES(CAST(:%s0 AS bigint))", prefix)
	for i := 1; i < length; i++ {
		orSlice[i] = fmt.Sprintf("(CAST(:%s%d AS bigint))", prefix, i)
	}

	sign := "IN"
	if !equal {
		sign = "NOT IN"
	}

	return fmt.Sprintf("%s %s (%s)", column, sign, strings.Join(orSlice, ", "))
}
func upgrade1to2(db *db.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(upgrade1To2MergeReadAndFav)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE users_articles_read")
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE users_articles_fav")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func upgrade2to3(db *db.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	t := time.Now().AddDate(0, -1, 0)
	_, err = tx.Exec(upgrade2To3SplitStatesToUnread, t)
	if err != nil {
		return err
	}

	_, err = tx.Exec(upgrade2To3SplitStatesToFavorite)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE users_articles_states")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func upgrade3to4(db *db.DB) error {
	table := "users_feeds_tags"
	var tabledef string

	for _, s := range initSQL {
		if strings.Contains(s, table) {
			tabledef = s
		}
	}

	if tabledef == "" {
		return errors.New("Definition for users_feeds_tags not found")
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("ALTER TABLE " + table + " RENAME TO " + table + "2")
	if err != nil {
		return err
	}

	_, err = tx.Exec(tabledef)
	if err != nil {
		return err
	}

	_, err = tx.Exec(upgrade3To4PopulateTags)
	if err != nil {
		return err
	}

	_, err = tx.Exec(upgrade3To4PopulateUsersFeedsTags)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE " + table + "2")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func init() {
	helper := &Helper{Helper: base.NewHelper()}

	helper.Set(db.SqlStmts{
		Feed: db.FeedStmts{AllForUser: getUserFeeds},
	})

	db.Register("postgres", helper)
}

const (
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = :user_login
ORDER BY f.title COLLATE "default"
`

	upgrade1To2MergeReadAndFav = `
INSERT INTO users_articles_states
SELECT COALESCE(ar.user_login, af.user_login), COALESCE(ar.article_id, af.article_id),
	CASE WHEN ar.article_id IS NULL THEN 'f'::BOOLEAN ELSE 't'::BOOLEAN END AS read,
	CASE WHEN af.article_id IS NULL THEN 'f'::BOOLEAN ELSE 't'::BOOLEAN END AS favorite
FROM users_articles_read ar FULL OUTER JOIN users_articles_fav af
	ON ar.article_id = af.article_id AND ar.user_login = af.user_login
ORDER BY ar.article_id
`

	upgrade2To3SplitStatesToUnread = `
INSERT INTO users_articles_unread (user_login, article_id, insert_date)
SELECT uas.user_login, uas.article_id, a.date
FROM users_articles_states uas INNER JOIN articles a
	ON uas.article_id = a.id
WHERE NOT uas.read AND a.date > $1
`

	upgrade2To3SplitStatesToFavorite = `
INSERT INTO users_articles_favorite (user_login, article_id)
SELECT uas.user_login, uas.article_id
FROM users_articles_states uas
WHERE uas.favorite
`

	upgrade3To4PopulateTags = `
INSERT INTO tags (value)
SELECT DISTINCT tag FROM users_feeds_tags2 WHERE tag != ''
`
	upgrade3To4PopulateUsersFeedsTags = `
INSERT INTO users_feeds_tags (user_login, feed_id, tag_id)
SELECT uft.user_login, uft.feed_id, t.id
FROM tags t INNER JOIN users_feeds_tags2 uft
	ON t.value = uft.tag
`
)
