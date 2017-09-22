package content_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"testing"

	"github.com/urandom/readeef/content"
	"golang.org/x/crypto/scrypt"
)

var (
	secret = []byte("secret")
)

func TestUser_Password(t *testing.T) {
	tests := []struct {
		name     string
		Login    content.Login
		password string
	}{
		{"simle", "test1", "password1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := content.User{
				Login: tt.Login,
			}
			if err := u.Password(tt.password, secret); err != nil {
				t.Errorf("User.Password() error = %v", err)
			}

			if !bytes.Equal(u.MD5API, md5Sum(string(u.Login), tt.password)) {
				t.Error("User.Password() md5sum not equal")
			}

			if !bytes.Equal(u.Hash, scryptHash(t, tt.password, u.Salt)) {
				t.Error("User.Password() hash not equal")
			}
		})
	}
}

func TestUser_Authenticate(t *testing.T) {
	tests := []struct {
		name          string
		Login         content.Login
		password      string
		passwordCheck string
		want          bool
	}{
		{"success", "test1", "password", "password", true},
		{"failure", "test2", "password", "password2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := content.User{
				Login: tt.Login,
			}
			u.Password(tt.password, secret)
			got, err := u.Authenticate(tt.passwordCheck, secret)
			if err != nil {
				t.Errorf("User.Authenticate() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("User.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Validate(t *testing.T) {
	type fields struct {
		Login content.Login
		Email string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"valid without email", fields{Login: "user1"}, false},
		{"valid with email", fields{Login: "user1", Email: "foo@sugr.org"}, false},
		{"invalid with email", fields{Login: "user1", Email: "foo"}, true},
		{"invalid without login", fields{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := content.User{
				Login: tt.fields.Login,
				Email: tt.fields.Email,
			}
			if err := u.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func md5Sum(login, password string) []byte {
	sum := md5.Sum([]byte(fmt.Sprintf("%s:%s", login, password)))
	return sum[:]
}

func scryptHash(t *testing.T, password string, salt []byte) []byte {
	dk, err := scrypt.Key([]byte(password), salt, 16384, 8, 1, 32)
	if err != nil {
		t.Fatal(err)
	}

	return dk
}
