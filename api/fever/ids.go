package fever

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func unreadItemIDs(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
	log.Infoln("Fetching unread fever item ids")

	ids := user.Ids(data.ArticleIdQueryOptions{UnreadOnly: true})
	if user.HasErr() {
		return errors.Wrap(user.Err(), "getting unread ids")
	}

	var buf bytes.Buffer

	for i := range ids {
		if i != 0 {
			buf.WriteString(",")
		}

		buf.WriteString(strconv.FormatInt(int64(ids[i]), 10))
	}

	resp["unread_item_ids"] = buf.String()

	return nil
}

func savedItemIDs(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
	log.Infoln("Fetching saved fever item ids")

	ids := user.Ids(data.ArticleIdQueryOptions{FavoriteOnly: true})
	if user.HasErr() {
		return errors.Wrap(user.Err(), "getting favorite ids")
	}

	var buf bytes.Buffer

	for i := range ids {
		if i != 0 {
			buf.WriteString(",")
		}

		buf.WriteString(strconv.FormatInt(int64(ids[i]), 10))
	}

	resp["saved_item_ids"] = buf.String()

	return nil
}
