package readeef

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
)
import "testing"

var file = "readeef-test.sqlite"
var conn = "file:./" + file + "?cache=shared&mode=rwc"

func TestUsers(t *testing.T) {
	os.Remove(file)

	db, err := NewDB("sqlite3", conn)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.GetUser("test"); err != nil {
		if err != sql.ErrNoRows {
			t.Fatal(err)
		}
	} else {
		t.Fatalf("Expected to get an error\n")
	}

	u := User{Login: "test", FirstName: "Hello", LastName: "World", Email: "test"}

	if err = db.UpdateUser(u); err == nil {
		t.Fatalf("Expected a validation error\n")
	} else {
		if _, ok := err.(ValidationError); !ok {
			t.Fatalf("Expected a validation error \n")
		}
	}

	u.Email = "test@example.com"

	if err = db.UpdateUser(u); err != nil {
		t.Fatal(err)
	}

	u, err = db.GetUser("test")
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

	err = u.setPassword("foobar")
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
		t.Fatal("Expected %s, got %s", u.Login, u2.Login)
	}

	if u.FirstName != u2.FirstName {
		t.Fatal("Expected %s, got %s", u.FirstName, u2.FirstName)
	}

	if u.LastName != u2.LastName {
		t.Fatal("Expected %s, got %s", u.LastName, u2.LastName)
	}

	if u.Email != u2.Email {
		t.Fatal("Expected %s, got %s", u.Email, u2.Email)
	}

	if !bytes.Equal(u.Salt, u2.Salt) {
		t.Fatal("Expected %s, got %s", u.Salt, u2.Salt)
	}

	if !bytes.Equal(u.Hash, u2.Hash) {
		t.Fatal("Expected %s, got %s", u.Hash, u2.Hash)
	}

	if !bytes.Equal(u.MD5API, u2.MD5API) {
		t.Fatal("Expected %s, got %s", u.MD5API, u2.MD5API)
	}

	db.Close()

	os.Remove(file)
}
