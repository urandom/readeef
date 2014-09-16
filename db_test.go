package readeef

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"math"
	"readeef/parser"
)
import (
	"testing"
	"time"
)

var db DB

func TestDBUsers(t *testing.T) {
	cleanDB(t)

	if _, err := db.GetUser("test"); err != nil {
		if err != sql.ErrNoRows {
			t.Fatal(err)
		}
	} else {
		t.Fatalf("Expected to get an error\n")
	}

	u := User{Login: "test", FirstName: "Hello", LastName: "World", Email: "test"}

	if err := db.UpdateUser(u); err == nil {
		t.Fatalf("Expected a validation error\n")
	} else {
		if _, ok := err.(ValidationError); !ok {
			t.Fatalf("Expected a validation error, got '%v'\n", err)
		}
	}

	u.Email = "test@example.com"

	if err := db.UpdateUser(u); err != nil {
		t.Fatal(err)
	}

	u, err := db.GetUser("test")
	if err != nil {
		t.Fatal(err)
	}

	if len(u.Salt) != 0 {
		t.Fatalf("Expected an empty u.Salt, got %s\n", u.Salt)
	}

	if len(u.Hash) != 0 {
		t.Fatalf("Expected an empty u.Hash, got %s\n", u.Hash)
	}

	if len(u.MD5API) != 0 {
		t.Fatalf("Expected an empty u.MD5API, got %s\n", u.MD5API)
	}

	err = u.SetPassword("foobar")
	if err != nil {
		t.Fatal(err)
	}

	if len(u.Salt) == 0 {
		t.Fatalf("Expected a non- empty u.Salt\n")
	}

	if len(u.Hash) == 0 {
		t.Fatalf("Expected a non-empty u.Hash\n")
	}

	if len(u.MD5API) == 0 {
		t.Fatalf("Expected a non-empty u.MD5API\n")
	}

	sum := md5.Sum([]byte(fmt.Sprintf("%s:%s", u.Login, "foobar")))
	if !bytes.Equal(u.MD5API, sum[:]) {
		t.Fatalf("Generated md5 sum doesn't equal the expected one\n")
	}

	if err = db.UpdateUser(u); err != nil {
		t.Fatal(err)
	}

	u2, err := db.GetUser("test")
	if err != nil {
		t.Fatal(err)
	}

	if u.Login != u2.Login {
		t.Fatalf("Expected %s, got %s", u.Login, u2.Login)
	}

	if u.FirstName != u2.FirstName {
		t.Fatalf("Expected %s, got %s", u.FirstName, u2.FirstName)
	}

	if u.LastName != u2.LastName {
		t.Fatalf("Expected %s, got %s", u.LastName, u2.LastName)
	}

	if u.Email != u2.Email {
		t.Fatalf("Expected %s, got %s", u.Email, u2.Email)
	}

	if !bytes.Equal(u.Salt, u2.Salt) {
		t.Fatalf("Expected %v, got %v", u.Salt, u2.Salt)
	}

	if !bytes.Equal(u.Hash, u2.Hash) {
		t.Fatalf("Expected %v, got %v", u.Hash, u2.Hash)
	}

	if !bytes.Equal(u.MD5API, u2.MD5API) {
		t.Fatalf("Expected %s, got %s", u.MD5API, u2.MD5API)
	}

	u.ProfileData["pi"] = math.Pi
	u.ProfileData["string_test"] = "Hello World"
	u.ProfileData["array"] = []int{1, 2, 3}

	if err := db.UpdateUser(u); err != nil {
		t.Fatal(err)
	}

	if u2, err := db.GetUser(u.Login); err == nil {
		if u2.ProfileData["pi"] != u.ProfileData["pi"] {
			t.Fatalf("Expected 'pi' to be '%s', got '%s'\n", u.ProfileData["pi"], u2.ProfileData["pi"])
		}
		if u2.ProfileData["string_test"] != u.ProfileData["string_test"] {
			t.Fatalf("Expected 'string_test' to be '%s', got '%s'\n", u.ProfileData["string_test"], u2.ProfileData["string_test"])
		}

		if len(u2.ProfileData["array"].([]interface{})) != 3 {
			t.Fatalf("Expected 'array' to be of size 3, got %d\n", len(u2.ProfileData["array"].([]int)))
		}
	} else {
		t.Fatal(err)
	}

	if err := db.DeleteUser(u); err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetUser("test"); err == nil || err != sql.ErrNoRows {
		t.Fatalf("Expected to not find the user\n")
	}
}

func TestDBFeeds(t *testing.T) {
	cleanDB(t)

	var err error

	if _, err := db.GetFeed(0); err != nil {
		if err != sql.ErrNoRows {
			t.Fatal(err)
		}
	} else {
		t.Fatalf("Expected to get an error\n")
	}

	f := Feed{Feed: parser.Feed{Title: "test"}}
	if f, _, err = db.UpdateFeed(f); err != nil {
		if _, ok := err.(ValidationError); !ok {
			t.Fatalf("Expected a validation error, got '%v'\n", err)
		}
	} else {
		t.Fatalf("Expected to get an error\n")
	}

	f.Link = "http://example.com"
	if f, _, err = db.UpdateFeed(f); err != nil {
		t.Fatal(err)
	}

	if f.Id == 0 {
		t.Fatal("Expected the feed to get a valid id after insertion in the db\n")
	}

	expectedStr := "http://example.com"
	if f, err = db.GetFeed(f.Id); err != nil {
		t.Fatal(err)
	} else {
		if f.Link != expectedStr {
			t.Fatalf("Expected '%s' for a link, got '%s'\n", expectedStr, f.Link)
		}
	}

	f.Title = "Example rss"
	if _, _, err = db.UpdateFeed(f); err != nil {
		t.Fatal(err)
	}

	expectedStr = "Example rss"
	if f, err := db.GetFeed(f.Id); err != nil {
		t.Fatal(err)
	} else {
		if f.Title != expectedStr {
			t.Fatalf("Expected '%s' for a title, got '%s'\n", expectedStr, f.Title)
		}
	}

	u := User{Login: "test"}
	if err := db.UpdateUser(u); err != nil {
		t.Fatal(err)
	}

	expectedInt := 0
	if feeds, err := db.GetUserFeeds(u); err == nil {
		if len(feeds) != expectedInt {
			t.Fatalf("Expected %d user feeds, got %d\n", expectedInt, len(feeds))
		}
	} else {
		t.Fatal(err)
	}

	f2 := Feed{Link: "http://rss.example.com"}
	if f2, _, err = db.UpdateFeed(f2); err != nil {
		t.Fatal(err)
	}

	expectedInt = 2
	if feeds, err := db.GetFeeds(); err == nil {
		if len(feeds) != expectedInt {
			t.Fatalf("Expected %d feeds, got %d\n", expectedInt, len(feeds))
		}
	} else {
		t.Fatal(err)
	}

	if f, err = db.CreateUserFeed(u, f); err != nil {
		t.Fatal(err)
	}

	if feeds, err := db.GetUserFeeds(u); err == nil {
		expectedInt = 1
		if len(feeds) != expectedInt {
			t.Fatalf("Expected %d user feeds, got %d\n", expectedInt, len(feeds))
		}
		expectedStr = "Example rss"
		if feeds[0].Title != expectedStr {
			t.Fatalf("Expected '%s' for a title, got '%s'\n", expectedStr, feeds[0].Title)
		}
	} else {
		t.Fatal(err)
	}

	if f2, err = db.CreateUserFeed(u, f2); err != nil {
		t.Fatal(err)
	}

	if feeds, err := db.GetUserFeeds(u); err == nil {
		expectedInt = 2
		if len(feeds) != expectedInt {
			t.Fatalf("Expected %d user feeds, got %d\n", expectedInt, len(feeds))
		}
		expectedStr = "Example rss"
		if feeds[1].Title != expectedStr {
			t.Fatalf("Expected '%s' for a title, got '%s'\n", expectedStr, feeds[1].Title)
		}
		expectedStr = f2.Link
		if feeds[0].Link != expectedStr {
			t.Fatalf("Expected '%s' for a link, got '%s'\n", expectedStr, feeds[0].Link)
		}
	} else {
		t.Fatal(err)
	}

	if err := db.DeleteUserFeed(f); err != nil {
		t.Fatal(err)
	}

	if feeds, err := db.GetUserFeeds(u); err == nil {
		expectedInt = 1
		if len(feeds) != expectedInt {
			t.Fatalf("Expected %d user feeds, got %d\n", expectedInt, len(feeds))
		}
		expectedStr = f2.Link
		if feeds[0].Link != expectedStr {
			t.Fatalf("Expected '%s' for a link, got '%s'\n", expectedStr, feeds[0].Link)
		}
	} else {
		t.Fatal(err)
	}

	f2.User = u
	if f, err := db.GetFeedArticles(f2); err == nil {
		expectedInt = 0
		if len(f.Articles) != expectedInt {
			t.Fatalf("Expected %d feed articles, got %d\n", expectedInt, len(f.Articles))
		}
	} else {
		t.Fatal(err)
	}

	t1, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
	if err != nil {
		t.Fatal(err)
	}

	a1 := Article{Article: parser.Article{Id: "1", Title: "Test 1"}, FeedId: f2.Id}
	a2 := Article{Article: parser.Article{Id: "2", Title: "Test 2", Date: t1}, FeedId: f2.Id}

	if f, err := db.CreateFeedArticles(f2, []Article{a1, a2}); err == nil {
		expectedInt = 2
		if len(f.Articles) != expectedInt {
			t.Fatalf("Expected %d feed articles, got %d\n", expectedInt, len(f.Articles))
		}
	} else {
		t.Fatal(err)
	}

	expectedDate := t1
	if f, err := db.GetFeedArticles(f2); err == nil {
		expectedInt = 2
		if len(f.Articles) != expectedInt {
			t.Fatalf("Expected %d feed articles, got %d\n", expectedInt, len(f.Articles))
		}
		expectedStr = "1"
		if f.Articles[0].Id != expectedStr {
			t.Fatalf("Expected '%s' for article id, got '%s'\n", expectedStr, f.Articles[0].Id)
		}
		expectedStr = "Test 2"
		if f.Articles[1].Title != expectedStr {
			t.Fatalf("Expected '%s' for article title, got '%s'\n", expectedStr, f.Articles[1].Title)
		}
		if !f.Articles[1].Date.Equal(expectedDate) {
			t.Fatalf("Expected '%s' for article date, got '%s'\n", expectedDate, f.Articles[1].Date)
		}
	} else {
		t.Fatal(err)
	}

	if f, err := db.GetUnreadFeedArticles(f2); err == nil {
		expectedInt = 2
		if len(f.Articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(f.Articles))
		}
	}

	if articles, err := db.GetUnreadUserArticles(f2.User); err == nil {
		expectedInt = 2
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}

		f2.Articles = articles
	}

	if err := db.MarkUserArticlesAsRead(f2.User, f2.Articles, true); err != nil {
		t.Fatal(err)
	}

	if f, err := db.GetUnreadFeedArticles(f2); err == nil {
		expectedInt = 0
		if len(f.Articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(f.Articles))
		}
	}

	if articles, err := db.GetUnreadUserArticles(f2.User); err == nil {
		expectedInt = 0
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
	}

	if err := db.MarkUserArticlesAsRead(f2.User, f2.Articles, false); err != nil {
		t.Fatal(err)
	}

	if articles, err := db.GetUnreadUserArticles(f2.User); err == nil {
		expectedInt = 2
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
	}

	a3 := Article{Article: parser.Article{Id: "3", Title: "Test 3", Date: time.Now().Add(time.Minute)}, FeedId: f2.Id}

	f2.Articles = append(f2.Articles, a3)
	if f, err := db.CreateFeedArticles(f2, []Article{a3}); err == nil {
		a1, a2, a3 = f.Articles[0], f.Articles[1], f.Articles[2]
	} else {
		t.Fatal(err)
	}

	if err := db.MarkUserArticlesByDateAsRead(f2.User, time.Now(), true); err != nil {
		t.Fatal(err)
	}

	if articles, err := db.GetUnreadUserArticles(f2.User); err == nil {
		expectedInt = 1
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}

		expectedStr = "Test 3"
		if articles[0].Title != expectedStr {
			t.Fatalf("Expected '%s' for article title, got '%s'\n", expectedStr, articles[0].Title)
		}
	}

	if articles, err := db.GetUserFavoriteArticles(u); err == nil {
		expectedInt = 0
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
	} else {
		t.Fatal(err)
	}

	if err := db.MarkUserArticlesAsFavorite(u, []Article{a3, a1}, true); err != nil {
		t.Fatal(err)
	}

	if articles, err := db.GetUserFavoriteArticles(u); err == nil {
		expectedInt = 2
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
		expectedStr = "Test 1"
		if articles[0].Title != expectedStr {
			t.Fatalf("Expected '%s' for article title, got '%s'\n", expectedStr, articles[0].Title)
		}
		expectedStr = "Test 3"
		if articles[1].Title != expectedStr {
			t.Fatalf("Expected '%s' for article title, got '%s'\n", expectedStr, articles[1].Title)
		}
	} else {
		t.Fatal(err)
	}

	if err := db.MarkUserArticlesAsFavorite(u, []Article{a1, a2}, false); err != nil {
		t.Fatal(err)
	}

	if articles, err := db.GetUserFavoriteArticles(u); err == nil {
		expectedInt = 1
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
		expectedStr = "Test 3"
		if articles[0].Title != expectedStr {
			t.Fatalf("Expected '%s' for article title, got '%s'\n", expectedStr, articles[0].Title)
		}
	} else {
		t.Fatal(err)
	}

	f3 := Feed{Link: "http://rss2.example.com"}
	if f3, _, err = db.UpdateFeed(f3); err != nil {
		t.Fatal(err)
	}

	if f3, err = db.CreateUserFeed(u, f3); err != nil {
		t.Fatal(err)
	}

	if f3, err = db.CreateFeedArticles(f3, []Article{a1}); err != nil {
		t.Fatal(err)
	}

	if articles, err := db.GetUnreadUserArticles(u); err == nil {
		expectedInt = 2
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
	} else {
		t.Fatal(err)
	}

	if err := db.MarkFeedArticlesByDateAsRead(f3, time.Now(), true); err != nil {
		t.Fatal(err)
	}

	if articles, err := db.GetUnreadUserArticles(u); err == nil {
		expectedInt = 1
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
		expectedInt64 := f2.Id
		if articles[0].FeedId != expectedInt64 {
			t.Fatalf("Expected '%d' for article feed id, got '%d'\n", expectedInt64, articles[0].FeedId)
		}
	} else {
		t.Fatal(err)
	}

	if articles, err := db.GetReadUserArticles(u); err == nil {
		expectedInt = 3
		if len(articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(articles))
		}
	} else {
		t.Fatal(err)
	}

	if f, err := db.GetReadFeedArticles(f3); err == nil {
		expectedInt = 1
		if len(f.Articles) != expectedInt {
			t.Fatalf("Expected %d unread feed articles, got %d\n", expectedInt, len(f.Articles))
		}
		expectedInt64 := f3.Id
		if f.Articles[0].FeedId != expectedInt64 {
			t.Fatalf("Expected '%d' for article feed id, got '%d'\n", expectedInt64, f.Articles[0].FeedId)
		}
	}
}

func cleanDB(t *testing.T) {
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	sql := []string{
		`DELETE FROM users`,
		`DELETE FROM feeds`,
	}

	for _, s := range sql {
		stmt, err := tx.Preparex(s)
		if err != nil {
			t.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec()
		if err != nil {
			t.Fatal(err)
		}

	}

	tx.Commit()
}
