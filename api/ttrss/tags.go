package ttrss

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

type categoriesContent []cat

type cat struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Unread  int64  `json:"unread"`
	OrderId int64  `json:"order_id"`
}

func getCategories(req request, user content.User, service repo.Service) (interface{}, error) {
	articleRepo := service.ArticleRepo()
	tagRepo := service.TagRepo()

	cContent := categoriesContent{}

	tags, err := tagRepo.ForUser(user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting user tags")
	}
	for _, tag := range tags {
		ids, err := tagRepo.FeedIDs(tag, user)
		if err != nil {
			return nil, errors.WithMessage(err, "getting tag feed ids")
		}

		count, err := articleRepo.Count(user,
			content.UnreadOnly, content.FeedIDs(ids),
			content.Filters(content.GetUserFilters(user)),
		)
		if err != nil {
			return nil, errors.WithMessage(err, "getting unread tag count")
		}

		if count > 0 || !req.UnreadOnly {
			cContent = append(cContent,
				cat{Id: strconv.FormatInt(int64(tag.ID), 10), Title: string(tag.Value), Unread: count},
			)
		}
	}

	count, err := articleRepo.Count(user,
		content.UnreadOnly, content.UntaggedOnly,
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting unread untagged count")
	}

	if count > 0 || !req.UnreadOnly {
		cContent = append(cContent,
			cat{Id: strconv.FormatInt(CAT_UNCATEGORIZED, 10), Title: "Uncategorized", Unread: count},
		)
	}

	count, err = articleRepo.Count(user,
		content.UnreadOnly, content.FavoriteOnly,
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting unread favorite count")
	}

	if count > 0 || !req.UnreadOnly {
		cContent = append(cContent,
			cat{Id: strconv.FormatInt(CAT_SPECIAL, 10), Title: "Special", Unread: count},
		)
	}

	return cContent, nil
}

func getLabels(req request, user content.User, service repo.Service) (interface{}, error) {
	return []interface{}{}, nil
}

func setArticleLabel(req request, user content.User, service repo.Service) (interface{}, error) {
	return genericContent{Status: "OK", Updated: 0}, nil
}

func init() {
	actions["getCategories"] = getCategories
	actions["getLabels"] = getLabels
	actions["setArticleLabel"] = setArticleLabel
}
