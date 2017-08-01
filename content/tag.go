package content

type TagID int64
type TagValue string

type Tag struct {
	ID    TagId    `json:"id"`
	Value TagValue `json:"value"`
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
