package repo

// Service provices access to the different content repositories.
type Service interface {
	UserRepo() User
	TagRepo() Tag
	FeedRepo() Feed
	SubscriptionRepo() Subscription
	ArticleRepo() Article
	ExtractRepo() Extract
	ThumbnailRepo() Thumbnail
	ScoresRepo() Scores
}
