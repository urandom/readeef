package repo_test

import (
	"sync"
	"testing"

	"github.com/urandom/readeef/content"
)

const (
	user1 = "user1"
	user2 = "user2"
)

var userSync sync.Once

func Test_userRepo_Get(t *testing.T) {
	skipTest(t)
	setupUser()

	tests := []struct {
		name    string
		login   content.Login
		wantErr bool
	}{
		{"valid1", user1, false},
		{"valid2", user2, false},
		{"invalid", content.Login("user3"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.UserRepo()
			got, err := r.Get(tt.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("userRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && got.Login != "" || !tt.wantErr && got.Login != tt.login {
				t.Errorf("userRepo.Get() = %v, want %v", got.Login, tt.login)
			}
		})
	}
}

func Test_userRepo_All(t *testing.T) {
	skipTest(t)
	setupUser()

	r := service.UserRepo()
	got, err := r.All()

	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Errorf("userRepo.All() expected 2, got %v", len(got))
		return
	}

	expected := 2
	for _, u := range got {
		if u.Login == user1 || u.Login == user2 {
			expected--
		}
	}

	if expected != 0 {
		t.Errorf("userRepo.All() expected both existing users, found %d", expected)
	}
}

func Test_userRepo_Delete(t *testing.T) {
	skipTest(t)
	setupUser()

	createUser("user3")

	r := service.UserRepo()
	got, err := r.All()

	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 3 {
		t.Errorf("userRepo.Delete() expected 2, got %v", len(got))
		return
	}

	err = r.Delete(content.User{Login: "user3"})
	if err != nil {
		t.Errorf("userRepo.Delete() error %v", err)
		return
	}

	_, err = r.Get("user3")
	if err == nil {
		t.Errorf("userRepo.Delete() expected validation error, got nil")
		return
	}

	if !content.IsNoContent(err) {
		t.Errorf("userRepo.Delete() expected validation error, got %v", err)
		return
	}
}

func createUser(login content.Login) {
	r := service.UserRepo()
	u := content.User{Login: login}

	if err := r.Update(u); err != nil {
		panic(err)
	}
}

func setupUser() {
	if skip {
		return
	}

	userSync.Do(func() {
		createUser(user1)
		createUser(user2)
	})
}
