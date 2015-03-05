package ql

import (
	_ "github.com/cznic/ql/driver"
	"github.com/urandom/readeef/db"
	"github.com/urandom/readeef/db/base"
)

type Helper struct {
	base.Helper
}

func (h Helper) InitSQL() []string {
	return initSQL
}

func init() {
	helper := &Helper{Helper: base.NewHelper()}

	helper.Set("create_user_article_read", createUserArticleRead)
	helper.Set("create_user_article_favorite", createUserArticleFavorite)
	helper.Set("create_article_scores", createArticleScores)

	helper.Set("create_feed", createFeed)
	helper.Set("update_feed", updateFeed)
	helper.Set("delete_feed", deleteFeed)
	helper.Set("create_feed_article", createFeedArticle)
	helper.Set("get_all_feed_articles", getAllFeedArticles)
	helper.Set("get_latest_feed_articles", getLatestFeedArticles)
	helper.Set("create_all_users_articles_read_by_feed_date", createAllUsersArticlesReadByFeedDate)
	helper.Set("delete_all_users_articles_read_by_feed_date", deleteAllUsersArticlesReadByFeedDate)
	helper.Set("create_user_feed_tag", createUserFeedTag)

	helper.Set("get_feed", getFeed)
	helper.Set("get_feed_by_link", getFeedByLink)
	helper.Set("get_feeds", getFeeds)
	helper.Set("get_unsubscribed_feeds", getUnsubscribedFeeds)

	helper.Set("get_user_tag_feeds", getUserTagFeeds)
	helper.Set("create_all_user_tag_articles_read_by_date", createAllUserTagArticlesByDate)
	helper.Set("delete_all_user_tag_articles_read_by_date", deleteAllUserTagArticlesByDate)

	helper.Set("create_user", createUser)
	helper.Set("get_user_feed", getUserFeed)
	helper.Set("create_user_feed", createUserFeed)
	helper.Set("get_user_feeds", getUserFeeds)
	helper.Set("get_user_article_count", getUserArticleCount)
	helper.Set("create_all_user_articles_read_by_date", createAllUserArticlesReadByDate)
	helper.Set("delete_all_user_articles_read_by_date", deleteAllUserArticlesReadByDate)
	helper.Set("create_newer_user_articles_read_by_date", createNewerUserArticlesReadByDate)
	helper.Set("delete_newer_user_articles_read_by_date", deleteNewerUserArticlesReadByDate)

	db.Register("ql", helper)
}
