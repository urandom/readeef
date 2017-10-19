package db

import "github.com/jmoiron/sqlx"

type Helper interface {
	SQL() SqlStmts
	InitSQL() []string

	CreateWithID(tx *sqlx.Tx, sql string, arg interface{}) (int64, error)
	Upgrade(db *DB, old, new int) error

	WhereMultipleORs(string, string, int, bool) string
}

type ArticleStmts struct {
	Create string
	Update string

	GetUserlessTemplate      string
	GetTemplate              string
	CountTemplate            string
	CountUserFeedsJoin       string
	StateReadColumn          string
	StateUnreadJoin          string
	StateFavoriteJoin        string
	GetIDsTemplate           string
	DeleteStaleUnreadRecords string
	GetScoreJoin             string
	GetUntaggedJoin          string

	ReadStateInsertTemplate     string
	ReadStateDeleteTemplate     string
	FavoriteStateInsertTemplate string
	FavoriteStateDeleteTemplate string
}

type ExtractStmts struct {
	Get    string
	Create string
	Update string
}

type FeedStmts struct {
	Get          string
	GetByLink    string
	GetForUser   string
	All          string
	AllForUser   string
	AllForTag    string
	Unsubscribed string

	IDs    string
	Create string
	Update string
	Delete string

	GetUsers       string
	Attach         string
	Detach         string
	CreateUserTag  string
	DeleteUserTags string
}

type ScoresStmts struct {
	Get    string
	Create string
	Update string
}

type SubscriptionStmts struct {
	GetForFeed string
	All        string

	Create string
	Update string
}

type TagStmts struct {
	Get            string
	GetByValue     string
	AllForUser     string
	AllForFeed     string
	Create         string
	GetUserFeedIDs string
	DeleteStale    string
}

type ThumbnailStmts struct {
	Get    string
	Create string
	Update string
}

type UserStmts struct {
	Get         string
	GetByMD5API string
	All         string

	Create string
	Update string
	Delete string
}

type SqlStmts struct {
	Article      ArticleStmts
	Extract      ExtractStmts
	Feed         FeedStmts
	Scores       ScoresStmts
	Subscription SubscriptionStmts
	Tag          TagStmts
	Thumbnail    ThumbnailStmts
	User         UserStmts
}

func Register(driver string, helper Helper) {
	if helper == nil {
		panic("No helper provided")
	}

	if _, ok := helpers[driver]; ok {
		panic("Helper " + driver + " already registered")
	}

	helpers[driver] = helper
}

// Can't recover from missing driver or statement, panic
func (db DB) SQL() SqlStmts {
	driver := db.DriverName()

	if h, ok := helpers[driver]; ok {
		return h.SQL()
	} else {
		panic("No helper registered for " + driver)
	}
}
