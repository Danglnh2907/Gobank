package main

import (
	"encoding/json"
	"fmt"
	"gobank/auth"
	"gobank/model"
	"gobank/user"
	"io"
	"net/http"
	"os"
	"strings"
)

func intializeDataFile() error {
	dirPath, filePath := "./data", "./data/credential.json"

	//Create data folder
	if _, err := os.Stat(dirPath); err != nil {
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			return err
		}
	}

	//Create credential.json file
	if _, err := os.Stat(filePath); err != nil {
		_, err = os.Create(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncData() error {
	//Check if client has been logged in
	data, err := os.ReadFile("./data/credential.json")
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	//Unmarshal credential
	var credential model.Credential
	err = json.Unmarshal(data, &credential)
	if err != nil {
		return err
	}

	//Make server called to fetch new credential data
	url := "http://localhost:8800/refresh"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("token", credential.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//Handle each respond status
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Failed to refresh credential from server")
		return nil
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotAcceptable {
		fmt.Println(string(data))
		auth.Logout()
		return nil
	}

	if resp.StatusCode == http.StatusFound {
		//Write data to credential.json
		err = os.WriteFile("./data/credential.json", data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func welcome() {
	//Read data from credential.json
	filePath := "./data/credential.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error at: welcome -> Error reading credential")
		fmt.Println(err)
		return
	}
	//Check if data is empty
	if len(data) == 0 {
		fmt.Println("Welcome to Gobank! Log in to continue with our service. Or create new account if you are new here")
		fmt.Println("Run './gobank help' for further assistant")
		return
	}
	//If data is not empty
	var credential model.Credential
	err = json.Unmarshal(data, &credential)
	if err != nil {
		fmt.Println("Error at: welcome -> Error unmarshal credential")
		fmt.Println(err)
		return
	}
	fmt.Printf("Welcome back, %s\n", credential.Info.Fullname)
}

func main() {
	//Intialize data folder
	err := intializeDataFile()
	if err != nil {
		fmt.Println("Error at: main -> Error intialize folders")
		fmt.Println(err)
		return
	}

	//Refresh credential every time user issue a command
	err = syncData()
	if err != nil {
		fmt.Println("Error at: main -> Error synchronize data from server")
		fmt.Println(err)
		return
	}

	//if len(args) == 1; call welcome()
	if len(os.Args) == 1 {
		welcome()
		return
	}

	/*----If len(args) != 1----*/

	//auth function (both user and admin)
	command := strings.ToLower(os.Args[1])

	if command == "register" || command == "regs" {
		if len(os.Args) == 2 {
			fmt.Println("Missing argument")
			return
		}

		if len(os.Args) == 3 {
			flag := strings.ToLower(os.Args[2])
			if flag == "--admin" {
				auth.Register("admin")
				return
			}

			if flag == "--user" {
				auth.Register("user")
				return
			}

			fmt.Println("Invalid argument")
			return
		}

		if len(os.Args) > 3 {
			fmt.Println("Too many arguments")
			return
		}
	}

	if command == "login" {
		if len(os.Args) == 2 {
			fmt.Println("Missing arguments")
			return
		}

		if len(os.Args) == 3 {
			flag := strings.ToLower(os.Args[2])
			if flag == "--admin" {
				auth.Login("admin")
				return
			}

			if flag == "--user" {
				auth.Login("user")
				return
			}

			fmt.Println("Invalid argument")
			return
		}

		if len(os.Args) > 3 {
			fmt.Println("Too many arguments")
			return
		}
	}

	if command == "update-password" || command == "upt-pass" {
		if len(os.Args) > 2 {
			fmt.Println("Too many arguments")
			return
		}

		auth.UpdatePassword()
		return
	}

	if command == "logout" {
		if len(os.Args) > 2 {
			fmt.Println("Too many arguments")
			return
		}
		auth.Logout()
		return
	}

	if command == "show-info" {
		if len(os.Args) > 2 {
			fmt.Println("Too many arguments")
			return
		}

		auth.ShowInfo()
		return
	}

	//user function
	if command == "topup" {
		if len(os.Args) > 2 {
			fmt.Println("Too many arguments")
			return
		}

		user.Topup()
		return
	}

	if command == "withdraw" {
		if len(os.Args) > 2 {
			fmt.Println("Too many arguments")
			return
		}

		user.Withdraw()
		return
	}

	if command == "make-transaction" || command == "mktrs" {
		if len(os.Args) > 2 {
			fmt.Println("Too many arguments")
			return
		}
		user.MakeTransaction()
		return
	}

	if command == "get-transactions" || command == "gettrs" {
		if len(os.Args) > 2 {
			fmt.Println("Too many arguments")
			return
		}
		user.GetTransactions()
		return
	}
	//admin function

	/*Unsupported command*/
}
