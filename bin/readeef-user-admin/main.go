package main

import (
	"flag"
	"fmt"
	"os"
	"readeef"
	"strings"
)

func main() {
	confpath := flag.String("config", "", "config path")

	flag.Parse()

	cfg, err := readeef.ReadConfig(*confpath)
	if err != nil {
		exitWithError(fmt.Sprintf("Error reading config from path '%s': %v", *confpath, err))
	}

	db := readeef.NewDB(cfg.DB.Driver, cfg.DB.Connect)
	if err := db.Connect(); err != nil {
		exitWithError(fmt.Sprintf("Error connecting to database: %v", err))
	}

	switch flag.Arg(0) {
	case "add":
		if flag.NArg() != 3 {
			exitWithError("Not enough arguments for 'add' command. Login and password must be specified")
		}
		login := flag.Arg(1)
		pass := flag.Arg(2)

		u := readeef.User{Login: login, Active: true}
		if err := u.SetPassword(pass); err != nil {
			exitWithError(fmt.Sprintf("Error setting password for user '%s': %v", login, err))
		}

		if err := db.UpdateUser(u); err != nil {
			exitWithError(fmt.Sprintf("Error updating the user database record for '%s': %v", login, err))
		}
	case "delete":
		if flag.NArg() != 2 {
			exitWithError("Not enough arguments for 'delete' command. Login must be specified")
		}
		login := flag.Arg(1)

		u, err := db.GetUser(login)
		if err != nil {
			exitWithError(fmt.Sprintf("Error getting user '%s' from the database: %v", login, err))
		}

		if err := db.DeleteUser(u); err != nil {
			exitWithError(fmt.Sprintf("Error removing user '%s' from the database: %v", login, err))
		}
	case "get":
		if flag.NArg() != 3 {
			exitWithError("Not enough arguments for 'get' command. Login and user property must be specified")
		}
		login := flag.Arg(1)
		prop := flag.Arg(2)

		u, err := db.GetUser(login)
		if err != nil {
			exitWithError(fmt.Sprintf("Error getting user '%s' from the database: %v", login, err))
		}

		lowerProp := strings.ToLower(prop)
		switch lowerProp {
		case "firstname", "first_name":
			fmt.Printf("%s\n", u.FirstName)
		case "lastname", "last_name":
			fmt.Printf("%s\n", u.LastName)
		case "email":
			fmt.Printf("%s\n", u.Email)
		case "hashtype", "hash_type":
			fmt.Printf("%s\n", u.HashType)
		case "salt":
			fmt.Printf("%v\n", u.Salt)
		case "hash":
			fmt.Printf("%v\n", u.Hash)
		case "md5api", "md5_api":
			fmt.Printf("%v\n", u.MD5API)
		case "admin":
			fmt.Printf("%v\n", u.Admin)
		case "active":
			fmt.Printf("%v\n", u.Active)
		default:
			exitWithError(fmt.Sprintf("Unknown user property '%s'", prop))
		}
	case "set":
		if flag.NArg() != 4 {
			exitWithError("Not enough arguments for 'update' command. Login, user property, and value must be specified")
		}
		login := flag.Arg(1)
		prop := flag.Arg(2)
		val := flag.Arg(3)

		u, err := db.GetUser(login)
		if err != nil {
			exitWithError(fmt.Sprintf("Error getting user '%s' from the database: %v", login, err))
		}

		lowerProp := strings.ToLower(prop)
		switch lowerProp {
		case "firstname", "first_name":
			u.FirstName = val
		case "lastname", "last_name":
			u.LastName = val
		case "email":
			u.Email = val
		case "password":
			if err := u.SetPassword(val); err != nil {
				exitWithError(fmt.Sprintf("Error setting password for user '%s': %v", u.Login, err))
			}
		case "admin", "active":
			enabled := false
			if val == "1" || val == "true" || val == "on" {
				enabled = true
			}
			if lowerProp == "admin" {
				u.Admin = enabled
			} else {
				u.Active = enabled
			}
		default:
			exitWithError(fmt.Sprintf("Unknown user property '%s'", prop))
		}

		if err := db.UpdateUser(u); err != nil {
			exitWithError(fmt.Sprintf("Error updating the user database record for '%s': %v", login, err))
		}
	case "list":
		users, err := db.GetUsers()
		if err != nil {
			exitWithError(fmt.Sprintf("Error getting users from the database: %v", err))
		}

		for _, u := range users {
			fmt.Printf("%s\n", u.Login)
		}
	case "list-detailed":
		users, err := db.GetUsers()
		if err != nil {
			exitWithError(fmt.Sprintf("Error getting users from the database: %v", err))
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
	default:
		exitWithError(fmt.Sprintf("Unknown command '%s'", flag.Arg(0)))
	}
}

func exitWithError(err string) {
	fmt.Fprintf(os.Stderr, err+"\n")
	os.Exit(1)
}
