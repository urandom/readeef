package readeef

const (
	get_user_feeds = ``
)

func (db DB) GetUserFeeds(u User) ([]Feed, error) {
	var f []Feed

	if err := db.Get(&f, get_user_feeds, u.Login); err != nil {
		return f, err
	}

	return f, nil
}
