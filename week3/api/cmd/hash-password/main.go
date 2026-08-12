package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func main() {
	fmt.Print("Password: ")

	password, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read password:", err)
		os.Exit(1)
	}

	password = strings.TrimRight(password, "\r\n")
	if password == "" {
		fmt.Fprintln(os.Stderr, "password must not be empty")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcryptCost,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to hash password:", err)
		os.Exit(1)
	}

	fmt.Println(string(hash))
}