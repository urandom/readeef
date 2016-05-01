package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
	_ "github.com/urandom/readeef/content/sql/db/postgres"
)

func main() {
	confpath := flag.String("config", "", "config path")

	flag.Parse()

	cfg, err := readeef.ReadConfig(*confpath)
	if err != nil {
		exitWithError(fmt.Sprintf("Error reading config from path '%s': %v", *confpath, err))
	}

	logger := readeef.NewLogger(cfg)
	repo, err := repo.New(cfg.DB.Driver, cfg.DB.Connect, logger)
	if err != nil {
		exitWithError(fmt.Sprintf("Error connecting to database: %v", err))
	}

	switch flag.Arg(0) {
	case "add":
		if flag.NArg() != 3 {
			exitWithError("Not enough arguments for 'add' command. Login and password must be specified")
		}
		login := flag.Arg(1)
		pass := flag.Arg(2)

		u := repo.User()
		i := u.Data()

		i.Login = data.Login(login)
		i.Active = true

		u.Data(i)

		u.Password(pass, []byte(cfg.Auth.Secret))
		u.Update()

		if u.HasErr() {
			exitWithError(fmt.Sprintf("Error setting password for user '%s': %v", login, u.Err()))
		}
	case "remove":
		if flag.NArg() != 2 {
			exitWithError("Not enough arguments for 'remove' command. Login must be specified")
		}
		login := flag.Arg(1)

		u := repo.UserByLogin(data.Login(login))
		u.Delete()

		if u.HasErr() {
			exitWithError(fmt.Sprintf("Error getting user '%s' from the database: %v", login, u.Err()))
		}
	case "get":
		if flag.NArg() != 3 {
			exitWithError("Not enough arguments for 'get' command. Login and user property must be specified")
		}
		login := flag.Arg(1)
		prop := flag.Arg(2)

		u := repo.UserByLogin(data.Login(login))
		if repo.HasErr() {
			exitWithError(fmt.Sprintf("Error getting user '%s' from the database: %v", login, repo.Err()))
		}

		lowerProp := strings.ToLower(prop)
		switch lowerProp {
		case "firstname", "first_name":
			fmt.Printf("%s\n", u.Data().FirstName)
		case "lastname", "last_name":
			fmt.Printf("%s\n", u.Data().LastName)
		case "email":
			fmt.Printf("%s\n", u.Data().Email)
		case "hashtype", "hash_type":
			fmt.Printf("%s\n", u.Data().HashType)
		case "salt":
			fmt.Printf("%v\n", u.Data().Salt)
		case "hash":
			fmt.Printf("%v\n", u.Data().Hash)
		case "md5api", "md5_api":
			fmt.Printf("%v\n", u.Data().MD5API)
		case "admin":
			fmt.Printf("%v\n", u.Data().Admin)
		case "active":
			fmt.Printf("%v\n", u.Data().Active)
		default:
			exitWithError(fmt.Sprintf("Unknown user property '%s'", prop))
		}
	case "set":
		if flag.NArg() != 4 {
			exitWithError("Not enough arguments for 'set' command. Login, user property, and value must be specified")
		}
		login := flag.Arg(1)
		prop := flag.Arg(2)
		val := flag.Arg(3)

		u := repo.UserByLogin(data.Login(login))
		if repo.HasErr() {
			exitWithError(fmt.Sprintf("Error getting user '%s' from the database: %v", login, repo.Err()))
		}

		in := u.Data()

		lowerProp := strings.ToLower(prop)
		switch lowerProp {
		case "firstname", "first_name":
			in.FirstName = val
		case "lastname", "last_name":
			in.LastName = val
		case "email":
			in.Email = val
		case "password":
			u.Password(val, []byte(cfg.Auth.Secret))
		case "admin", "active":
			enabled := false
			if val == "1" || val == "true" || val == "on" {
				enabled = true
			}
			if lowerProp == "admin" {
				in.Admin = enabled
			} else {
				in.Active = enabled
			}
		default:
			exitWithError(fmt.Sprintf("Unknown user property '%s'", prop))
		}

		u.Update()
		if u.HasErr() {
			exitWithError(fmt.Sprintf("Error updating the user database record for '%s': %v", login, u.Err()))
		}
	case "list":
		users := repo.AllUsers()
		if repo.HasErr() {
			exitWithError(fmt.Sprintf("Error getting users from the database: %v", repo.Err()))
		}

		for _, u := range users {
			fmt.Printf("%s\n", u.Data().Login)
		}
	case "list-detailed":
		users := repo.AllUsers()
		if repo.HasErr() {
			exitWithError(fmt.Sprintf("Error getting users from the database: %v", repo.Err()))
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
	default:
		exitWithError(fmt.Sprintf("Unknown command '%s'", flag.Arg(0)))
	}
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\t%s [arguments] command\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, `Available commands:

	add LOGIN PASSWORD 		add a new user
	remove LOGIN 			remove an existing user
	get LOGIN PROPERTY 		retrieve a property of a user
		- firstname 			the first name
		- lastname 			the last name
		- email 			the email
		- hashtype 			the password hash type
		- salt 				the salt
		- hash 				the password hash
		- md5api 			the password-based key
						used by the fever api emulation
		- admin 			whether the user is an admin
		- active 			whether the user is active
	set LOGIN PROPERTY VALUE 	sets a new value to a given property
		- instead of a salt/hashtype/hash and md5api, a password property is
		used
	list 				lists all users
	list-detailed 			lists all users, including some properties

`)
		fmt.Fprintf(os.Stderr, "Available arguments:\n\n")
		flag.PrintDefaults()
	}
}
