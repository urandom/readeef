package content

import "errors"

type TagID int64
type TagValue string

type Tag struct {
	ID    TagID    `json:"id"`
	Value TagValue `json:"value"`
}

func (t Tag) Validate() error {
	if t.Value == "" {
		return NewValidationError(errors.New("Tag has no value"))
	}

	return nil
}

func (t Tag) String() string {
	return string(t.Value)
}

/*
type Tag interface {
	Error
	ArticleSearch
	ArticleRepo
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Data(data ...data.Tag) data.Tag

	Validate() error

	AllFeeds() []TaggedFeed
}
*/
