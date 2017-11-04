package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/repo/sql"
	"github.com/urandom/readeef/log"
)

var (
	userAdminCommands = map[string]func([]string, repo.Service, config.Config, log.Log) error{}
)

func runUserAdmin(config config.Config, args []string) error {
	if len(args) == 0 {
		return errors.New("no command")
	}

	log := initLog(config.Log)
	service, err := sql.NewService(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return errors.WithMessage(err, "creating content service")
	}

	if f, ok := userAdminCommands[args[0]]; ok {
		return f(args[1:], service, config, log)
	}

	return errors.Errorf("unknown command %s", args[0])
}

func userAdminAdd(args []string, service repo.Service, config config.Config, log log.Log) error {
	if len(args) != 2 {
		return errors.New("invalid number of arguments")
	}

	u := content.User{}

	u.Login = content.Login(args[0])
	u.Active = true

	err := u.Password(args[1], []byte(config.Auth.Secret))
	if err == nil {
		err = service.UserRepo().Update(u)
	}

	if err != nil {
		return errors.WithMessage(err, "creating admin user")
	}

	return nil
}

func userAdminRemove(args []string, service repo.Service, config config.Config, log log.Log) error {
	if len(args) != 1 {
		return errors.New("invalid number of arguments")
	}

	repo := service.UserRepo()

	u, err := repo.Get(content.Login(args[0]))
	if err == nil {
		err = repo.Delete(u)
	}

	if err != nil {
		return errors.WithMessage(err, "deleting user")
	}

	return nil
}

func userAdminGet(args []string, service repo.Service, config config.Config, log log.Log) error {
	if len(args) != 2 {
		return errors.New("invalid number of arguments")
	}

	u, err := service.UserRepo().Get(content.Login(args[0]))
	if err != nil {
		return errors.WithMessage(err, "getting user")
	}

	prop := strings.ToLower(args[1])
	switch prop {
	case "firstname":
		fmt.Println(u.FirstName)
	case "lastname":
		fmt.Println(u.LastName)
	case "email":
		fmt.Println(u.Email)
	case "hashtype":
		fmt.Println(u.HashType)
	case "salt":
		fmt.Println(u.Salt)
	case "hash":
		fmt.Println(u.Hash)
	case "md5api":
		fmt.Println(u.MD5API)
	case "admin":
		fmt.Println(u.Admin)
	case "active":
		fmt.Println(u.Active)
	case "profile":
		if b, err := json.Marshal(u.ProfileData); err == nil {
			fmt.Println(string(b))
		} else {
			return errors.Wrapf(err, "marshaling %s profile data", u)
		}
	default:
		return errors.Errorf("unknown property %s", args[1])
	}

	return nil
}

func userAdminSet(args []string, service repo.Service, config config.Config, log log.Log) error {
	if len(args) != 3 {
		return errors.New("invalid number of arguments")
	}

	repo := service.UserRepo()
	u, err := repo.Get(content.Login(args[0]))
	if err != nil {
		return errors.WithMessage(err, "getting user")
	}

	prop := strings.ToLower(args[1])
	switch prop {
	case "firstname":
		u.FirstName = args[2]
	case "lastname":
		u.LastName = args[2]
	case "email":
		u.Email = args[2]
	case "admin", "active":
		enabled := false
		if args[2] == "1" || args[2] == "true" || args[2] == "on" {
			enabled = true
		}
		if prop == "admin" {
			u.Admin = enabled
		} else {
			u.Active = enabled
		}
	case "profile":
		if err = json.Unmarshal([]byte(args[2]), &u.ProfileData); err != nil {
			return errors.Wrapf(err, "unmarshaling profile data for %s", u)
		}
	case "password":
		if err = u.Password(args[2], []byte(config.Auth.Secret)); err != nil {
			return errors.Wrapf(err, "setting %s password", u)
		}
	default:
		return errors.Errorf("unknown property %s", args[1])
	}

	if err = repo.Update(u); err != nil {
		return errors.WithMessage(err, "updating user")
	}

	return nil
}

func userAdminList(args []string, service repo.Service, config config.Config, log log.Log) error {
	users, err := service.UserRepo().All()
	if err != nil {
		return errors.WithMessage(err, "getting all users")
	}

	for _, u := range users {
		fmt.Println(u.Login)
	}

	return nil
}

func userAdminListDetailed(args []string, service repo.Service, config config.Config, log log.Log) error {
	users, err := service.UserRepo().All()
	if err != nil {
		return errors.WithMessage(err, "getting all users")
	}

	for _, u := range users {
		fmt.Printf("Login: %s", u.Login)
		if u.FirstName != "" {
			fmt.Printf(", first name: %s", u.FirstName)
		}
		if u.LastName != "" {
			fmt.Printf(", last name: %s", u.LastName)
		}
		if u.Email != "" {
			fmt.Printf(", email: %s", u.Email)
		}
		if u.HashType != "" {
			fmt.Printf(", has type: %s", u.HashType)
		}
		fmt.Printf("\n")
	}

	return nil
}

func init() {
	flags := flag.NewFlagSet("user-admin", flag.ExitOnError)

	commands = append(commands, Command{
		Name:  "user-admin",
		Desc:  "create, delete and modify users",
		Flags: flags,
		Run:   runUserAdmin,
	})

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of user-admin:\n\n")
		fmt.Fprintf(os.Stderr, "\tuser-admin [arguments] command\n\n")
		fmt.Fprintf(os.Stderr, `Available commands:

	add LOGIN PASSWORD 		add a new user
	remove LOGIN 			remove an existing user
	get LOGIN PROPERTY 		retrieve a property of a user
		- firstname 		the first name
		- lastname 		the last name
		- email 		the email
		- hashtype 		the password hash type
		- salt 			the salt
		- hash 			the password hash
		- md5api 		the password-based key
					used by the fever api emulation
		- admin 		whether the user is an admin
		- active 		whether the user is active
		- profile 		the json profile data
	set LOGIN PROPERTY VALUE 	sets a new value to a given property
		- instead of a salt/hashtype/hash and md5api, a password property is used
	list 				lists all users
	list-detailed 			lists all users, including some properties

`)
	}

	userAdminCommands["add"] = userAdminAdd
	userAdminCommands["remove"] = userAdminRemove
	userAdminCommands["get"] = userAdminGet
	userAdminCommands["set"] = userAdminSet
	userAdminCommands["list"] = userAdminList
	userAdminCommands["list-detailed"] = userAdminListDetailed
}
