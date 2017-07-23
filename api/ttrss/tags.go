package ttrss

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type categoriesContent []cat

type cat struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Unread  int64  `json:"unread"`
	OrderId int64  `json:"order_id"`
}

func getCategories(req request, user content.User) (interface{}, error) {
	cContent := categoriesContent{}
	o := data.ArticleCountOptions{UnreadOnly: true}

	for _, t := range user.Tags() {
		td := t.Data()
		count := t.Count(o)

		if count > 0 || !req.UnreadOnly {
			cContent = append(cContent,
				cat{Id: strconv.FormatInt(int64(td.Id), 10), Title: string(td.Value), Unread: count},
			)
		}
	}

	count := user.Count(data.ArticleCountOptions{UnreadOnly: true, UntaggedOnly: true})
	if count > 0 || !req.UnreadOnly {
		cContent = append(cContent,
			cat{Id: strconv.FormatInt(TTRSS_CAT_UNCATEGORIZED, 10), Title: "Uncategorized", Unread: count},
		)
	}

	o.FavoriteOnly = true
	count = user.Count(o)

	if count > 0 || !req.UnreadOnly {
		cContent = append(cContent,
			cat{Id: strconv.FormatInt(TTRSS_CAT_SPECIAL, 10), Title: "Special", Unread: count},
		)
	}

	if user.HasErr() {
		return nil, errors.Wrapf(user.Err(), "getting user %s tags", user.Data().Login)
	}

	return cContent, nil
}

func getLabels(req request, user content.User) (interface{}, error) {
	return []interface{}{}, nil
}

func setArticleLabel(req request, user content.User) (interface{}, error) {
	return genericContent{Status: "OK", Updated: 0}, nil
}

func init() {
	actions["getCategories"] = getCategories
	actions["getLabels"] = getLabels
	actions["setArticleLabel"] = setArticleLabel
}
