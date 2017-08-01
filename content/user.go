package content

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"fmt"
	"net/mail"

	"golang.org/x/crypto/scrypt"

	"github.com/pkg/errors"
)

// Login is the user login name.
type Login string

// User represents a readeef user.
type User struct {
	Login       Login  `json:"login"`
	FirstName   string `db:"first_name" json:"firstName"`
	LastName    string `db:"last_name" json:"lastName"`
	Email       string `json:"email"`
	HashType    string `db:"hash_type" json:"-"`
	Admin       bool   `json:"admin"`
	Active      bool   `json:"active"`
	ProfileJSON []byte `db:"profile_data" json:"-"`
	Salt        []byte `json:"-"`
	Hash        []byte `json:"-"`
	MD5API      []byte `db:"md5_api" json:"-"` // "md5(user:pass)"

	ProfileData map[string]interface{}
}

func (u *User) Password(password string, secret []byte) error {
	h := md5.Sum([]byte(fmt.Sprintf("%s:%s", u.data.Login, password)))

	u.data.MD5API = h[:]

	c := 30
	salt := make([]byte, c)
	if _, err := rand.Read(salt); err != nil {
		return errors.Wrap(err, "generating salt")
	}

	u.data.Salt = salt
	u.data.HashType = "scrypt"
	hash, err := u.generateHash(password, secret)
	if err != nil {
		return errors.WithMessage(err, "generating password hash")
	}

	u.data.Hash = hash
}

// Validate checks whether all required fields have been provided.
func (u User) Validate() error {
	if u.data.Login == "" {
		return NewValidationError(errors.New("invalid user login"))
	}
	if u.data.Email != "" {
		if _, err := mail.ParseAddress(u.String()); err != nil {
			return NewValidationError(err)
		}
	}

	return nil
}

func (u User) Authenticate(password string, secret []byte) (bool, error) {
	if u.HasErr() {
		return false
	}

	hash, err := u.generateHash(password, secret)
	if err != nil {
		return false, errors.WithMessage(err, "authenticating user")
	}

	return subtle.ConstantTimeCompare(u.data.Hash, hash) == 1, nil
}

func (u User) String() string {
	if u.FirstName != "" && u.LastName != "" && u.Email != "" {
		return fmt.Sprintf("%s: %s %s (%s)", u.Login, u.FirstName, u.LastName, u.Email)
	} else if u.Email != "" {
		return fmt.Sprintf("%s: (%s)", u.Login, u.Email)
	} else {
		return u.Login
	}
}

func (u User) generateHash(password string, secret []byte) ([]byte, error) {
	switch u.HashType {
	case "sha1":
		hash := sha1.Sum(append(secret, append(u.Salt, []byte(password)...)...))

		return hash[:], nil
	case "scrypt":
		dk, err := scrypt.Key([]byte(password), u.Salt, 16384, 8, 1, 32)
		if err != nil {
			err = errors.Wrap(err, "generating scrypt key")
		}

		return dk, err
	default:
		panic("Unknown hash type " + u.HashType)
	}
}

/*
type User interface {
	Error
	RepoRelated
	ArticleSearch
	ArticleRepo

	fmt.Stringer
	json.Marshaler

	Data(data ...data.User) data.User

	Validate() error

	Password(password string, secret []byte)
	Authenticate(password string, secret []byte) bool

	Update()
	Delete()

	FeedById(id data.FeedId) UserFeed
	AddFeed(feed Feed) UserFeed

	AllFeeds() []UserFeed

	AllTaggedFeeds() []TaggedFeed

	ArticleById(id data.ArticleId, opts ...data.ArticleQueryOptions) UserArticle
	ArticlesById(ids []data.ArticleId, opts ...data.ArticleQueryOptions) []UserArticle

	Tags() []Tag
	TagById(id data.TagId) Tag
	TagByValue(v data.TagValue) Tag
}

type UserRelated interface {
	User(u ...User) User
}

type TokenStorage interface {
	Store(token string, expiration time.Time) error
	Exists(token string) (bool, error)
	RemoveExpired() error
}
*/
