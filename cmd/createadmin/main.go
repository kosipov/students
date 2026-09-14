// Command createadmin creates a user who can sign in to the admin panel.
//
// It uses the same MYSQL_* env variables and config/config.yml (auth.hash_salt)
// as the API, so it must be run from the app working directory, e.g. inside the container:
//
//	./create-admin -username admin
//
// The password is read from ADMIN_PASSWORD, otherwise prompted without echo
// (or read from stdin when it is not a terminal). It is never accepted as a flag
// so that it does not leak into shell history or the process list.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/kosipov/students/auth"
	"github.com/kosipov/students/auth/usecase"
	"github.com/kosipov/students/config"
	"github.com/kosipov/students/server"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"log"
	"os"
	"strings"
	"time"

	authgorm "github.com/kosipov/students/auth/repository/gorm"
)

const minPasswordLength = 8

func main() {
	username := flag.String("username", "", "admin login")
	flag.Parse()

	login := strings.TrimSpace(*username)
	if login == "" {
		flag.Usage()
		os.Exit(2)
	}

	password, err := readPassword()
	if err != nil {
		log.Fatalf("Failed to read password: %s", err)
	}
	if len(password) < minPasswordLength {
		log.Fatalf("Password must be at least %d characters long", minPasswordLength)
	}

	if err := config.Init(); err != nil {
		log.Fatalf("%s", err.Error())
	}

	db := server.InitDB()
	defer db.Close()

	authUC := usecase.NewAuthUseCase(
		authgorm.NewUserRepository(db),
		viper.GetString("auth.hash_salt"),
		[]byte(viper.GetString("auth.signing_key")),
		viper.GetDuration("auth.token_ttl"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = authUC.SignUp(ctx, login, password)
	if errors.Is(err, auth.ErrUserAlreadyExists) {
		log.Fatalf("User %q already exists", login)
	}
	if err != nil {
		log.Fatalf("Failed to create user: %s", err)
	}

	fmt.Printf("User %q created\n", login)
}

func readPassword() (string, error) {
	if password, ok := os.LookupEnv("ADMIN_PASSWORD"); ok {
		return password, nil
	}

	stdin := int(os.Stdin.Fd())
	if !term.IsTerminal(stdin) {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	fmt.Fprint(os.Stderr, "Password: ")
	password, err := term.ReadPassword(stdin)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}

	fmt.Fprint(os.Stderr, "Repeat password: ")
	confirmation, err := term.ReadPassword(stdin)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}

	if string(password) != string(confirmation) {
		return "", errors.New("passwords do not match")
	}
	return string(password), nil
}
