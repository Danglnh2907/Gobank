package main

import (
	//Import standard library
	"fmt"
	"log"
	"net/http"

	//Import user's defined package
	"gobank/auth"
	"gobank/user"
	"gobank/utility"
)

func main() {
	//Connect to database
	_, err := utility.ConnectDB("gobank")
	if err != nil {
		fmt.Println("Error at: main -> Error connecting to database")
		fmt.Println(err)
		return
	}

	//Set up the initial table in database
	utility.InitializeTable()

	//Setup mux and handle function
	mux := http.NewServeMux()

	//mux for auth (both user and admin)
	mux.HandleFunc("/register", auth.Register)
	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/refresh", auth.SendCredential)
	mux.HandleFunc("/update-password", auth.ChangePassword)

	//mux for user
	mux.HandleFunc("/topup", user.Topup)
	mux.HandleFunc("/withdraw", user.Withdraw)
	mux.HandleFunc("/fullname", user.GetFullname) //Find account's fullname based on account number
	mux.HandleFunc("/transaction", user.MakeTransaction)
	mux.HandleFunc("/transactions", user.GetTransactions)

	//Start server
	fmt.Println("Server start at http://localhost:8800")
	err = http.ListenAndServe("localhost:8800", mux)
	if err != nil {
		fmt.Println("Error at main -> Error starting server")
		log.Fatal(err)
	}
}
