package fever

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/pool"
)

func unreadItemIDs(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	log readeef.Logger,
) error {
	log.Infoln("Fetching unread fever item ids")

	ids, err := service.ArticleRepo().IDs(user, content.UnreadOnly)
	if err != nil {
		return errors.WithMessage(err, "getting unread ids")
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	for i := range ids {
		if i != 0 {
			buf.WriteString(",")
		}

		buf.WriteString(strconv.FormatInt(int64(ids[i]), 10))
	}

	resp["unread_item_ids"] = buf.String()

	return nil
}

func savedItemIDs(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	log readeef.Logger,
) error {
	log.Infoln("Fetching saved fever item ids")

	ids, err := service.ArticleRepo().IDs(user, content.UnreadOnly)
	if err != nil {
		return errors.WithMessage(err, "getting unread ids")
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	for i := range ids {
		if i != 0 {
			buf.WriteString(",")
		}

		buf.WriteString(strconv.FormatInt(int64(ids[i]), 10))
	}

	resp["saved_item_ids"] = buf.String()

	return nil
}

func init() {
	actions["unread_item_ids"] = unreadItemIDs
	actions["saved_item_ids"] = savedItemIDs
}
