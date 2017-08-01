package repo

// Service provices access to the different content repositories.
type Service interface {
	UserRepo() UserRepo
	FeedRepo() FeedRepo
	ArticleRepo() ArticleRepo
}
