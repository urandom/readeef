package token

import (
	"encoding/binary"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type BoltStorage struct {
	db *bolt.DB
}

var (
	bucket = []byte("token-bucket")
)

func NewBoltStorage(path string) (BoltStorage, error) {
	db, err := bolt.Open(path, 0660, nil)

	if err != nil {
		return BoltStorage{}, errors.Wrapf(err, "opening token bolt storage %s", path)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)

		return err
	})
	if err != nil {
		return BoltStorage{}, errors.Wrap(err, "creating token bolt bucket")
	}

	return BoltStorage{db}, nil

}

func (b BoltStorage) Store(token string, expiration time.Time) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)

		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(expiration.Unix()))

		return b.Put([]byte(token), buf)
	})

	if err != nil {
		err = errors.Wrap(err, "writing token to storage")
	}

	return err
}

func (b BoltStorage) Exists(token string) (bool, error) {
	exists := false

	err := b.db.View(func(tx *bolt.Tx) error {
		if v := tx.Bucket(bucket).Get([]byte(token)); v != nil {
			exists = true
		}

		return nil
	})

	if err != nil {
		err = errors.Wrap(err, "looking up token")
	}

	return exists, err
}

func (b BoltStorage) RemoveExpired() error {
	now := time.Now()

	err := b.db.Update(func(tx *bolt.Tx) error {

		c := tx.Bucket(bucket).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			t := time.Unix(int64(binary.LittleEndian.Uint64(v)), 0)

			if now.Before(t) {
				continue
			}

			if err := c.Delete(); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		err = errors.Wrap(err, "cleaning expired tokens")
	}

	return err
}
