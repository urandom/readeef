package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
)

var (
	userAdminCommands = map[string]func([]string, content.Repo, config.Config, readeef.Logger) error{}
)

func runUserAdmin(config config.Config, args []string) error {
	log := readeef.NewLogger(config.Log)
	repo, err := repo.New(config.DB.Driver, config.DB.Connect, log)
	if err != nil {
		return errors.WithMessage(err, "creating content repo")
	}

	if len(args) == 0 {
		return errors.New("no command")
	}

	if f, ok := userAdminCommands[args[0]]; ok {
		return f(args[1:], repo, config, log)
	}

	return errors.Errorf("unknown command %s", args[0])
}

func userAdminAdd(args []string, repo content.Repo, config config.Config, log readeef.Logger) error {
	if len(args) != 2 {
		return errors.New("invalid number of arguments")
	}

	u := repo.User()
	i := u.Data()

	i.Login = data.Login(args[0])
	i.Active = true

	u.Data(i)
	u.Password(args[1], []byte(config.Auth.Secret))

	if u.HasErr() {
		return errors.WithMessage(u.Err(), "creating user")
	}

	return nil
}

func userAdminRemove(args []string, repo content.Repo, config config.Config, log readeef.Logger) error {
	if len(args) != 1 {
		return errors.New("invalid number of arguments")
	}

	u := repo.UserByLogin(data.Login(args[0]))
	u.Delete()

	if u.HasErr() {
		return errors.WithMessage(u.Err(), "deleting user")
	}

	return nil
}

func userAdminGet(args []string, repo content.Repo, config config.Config, log readeef.Logger) error {
	if len(args) != 2 {
		return errors.New("invalid number of arguments")
	}

	u := repo.UserByLogin(data.Login(args[0]))
	if u.HasErr() {
		return errors.WithMessage(u.Err(), "getting user")
	}

	prop := strings.ToLower(args[1])
	switch prop {
	case "firstname":
		fmt.Println(u.Data().FirstName)
	case "lastname":
		fmt.Println(u.Data().LastName)
	case "email":
		fmt.Println(u.Data().Email)
	case "hashtype":
		fmt.Println(u.Data().HashType)
	case "salt":
		fmt.Println(u.Data().Salt)
	case "hash":
		fmt.Println(u.Data().Hash)
	case "md5api":
		fmt.Println(u.Data().MD5API)
	case "admin":
		fmt.Println(u.Data().Admin)
	case "active":
		fmt.Println(u.Data().Active)
	default:
		return errors.Errorf("unknown property %s", args[1])
	}

	return nil
}

func userAdminSet(args []string, repo content.Repo, config config.Config, log readeef.Logger) error {
	if len(args) != 3 {
		return errors.New("invalid number of arguments")
	}

	u := repo.UserByLogin(data.Login(args[0]))
	if u.HasErr() {
		return errors.WithMessage(u.Err(), "getting user")
	}

	in := u.Data()

	prop := strings.ToLower(args[1])
	switch prop {
	case "firstname":
		in.FirstName = args[2]
	case "lastname":
		in.LastName = args[2]
	case "email":
		in.Email = args[2]
	case "admin", "active":
		enabled := false
		if args[2] == "1" || args[2] == "true" || args[2] == "on" {
			enabled = true
		}
		if prop == "admin" {
			in.Admin = enabled
		} else {
			in.Active = enabled
		}
	case "password":
		u.Password(args[2], []byte(config.Auth.Secret))
	default:
		return errors.Errorf("unknown property %s", args[1])
	}

	u.Data(in)
	u.Update()

	if u.HasErr() {
		return errors.WithMessage(u.Err(), "updating user")
	}

	return nil
}

func userAdminList(args []string, repo content.Repo, config config.Config, log readeef.Logger) error {
	users := repo.AllUsers()
	if repo.HasErr() {
		return errors.WithMessage(repo.Err(), "getting all users")
	}

	for _, u := range users {
		fmt.Println(u.Data().Login)
	}

	return nil
}

func userAdminListDetailed(args []string, repo content.Repo, config config.Config, log readeef.Logger) error {
	users := repo.AllUsers()
	if repo.HasErr() {
		return errors.WithMessage(repo.Err(), "getting all users")
	}

	for _, u := range users {
		in := u.Data()
		fmt.Printf("Login: %s", in.Login)
		if in.FirstName != "" {
			fmt.Printf(", first name: %s", in.FirstName)
		}
		if in.LastName != "" {
			fmt.Printf(", last name: %s", in.LastName)
		}
		if in.Email != "" {
			fmt.Printf(", email: %s", in.Email)
		}
		if in.HashType != "" {
			fmt.Printf(", has type: %s", in.HashType)
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
