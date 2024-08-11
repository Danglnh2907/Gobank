package user

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gobank/auth"
	"gobank/model"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func Topup() {
	//Check if client has been logged in
	data, err := os.ReadFile("./data/credential.json")
	if err != nil {
		fmt.Println("Error at: Topup -> Error reading credential")
		fmt.Println(err)
		return
	}

	if len(data) == 0 {
		fmt.Println("You haven't logged in! This service required you to log in to continue")
		return
	}

	//Get token from credential
	var credential model.Credential
	err = json.Unmarshal(data, &credential)
	if err != nil {
		fmt.Println("Error at: Topup -> Error unmarshal crdential")
		fmt.Println(err)
		return
	}
	token := credential.Token

	var (
		amount  float64
		temp    string
		isValid bool
		reader  = bufio.NewReader(os.Stdin)
	)

	//Ask for user's amount of money
	isValid = false
	for !isValid {
		//Read amount from stdin
		fmt.Print("Enter your amount of money: ")
		temp, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: Topup -> Error reading amount from stdin")
			fmt.Println(err)
			return
		}
		temp = strings.TrimSpace(temp)
		amount, err = strconv.ParseFloat(temp, 64)
		if err != nil {
			fmt.Println("Invalid value for amount!")
			continue
		}

		//Check if amount is valid number
		isValid = amount > 0
		if !isValid {
			fmt.Println("The amount of money must be greater than 0")
		}
	}

	//Package data before sending to server
	data, err = json.MarshalIndent(amount, "", " ")
	if err != nil {
		fmt.Println("Error at: Topup -> Error marshal data")
		fmt.Println(err)
		return
	}

	//Make new request to server
	url := "http://localhost:8800/topup"
	req, err := http.NewRequest("UPDATE", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error at: Topup -> Error making new request to server")
		fmt.Println(err)
		return
	}
	req.Header.Set("token", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at: Topup -> Error sending request to server or failed to receive respond")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	//Handle each respond status
	message, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error at: Topup -> Error reading respond body")
		fmt.Println(err)
		return
	}

	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error :(")
		return
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotAcceptable {
		fmt.Println(string(message))
		auth.Logout()
		return
	}

	if resp.StatusCode == http.StatusOK {
		//Update balance in credential
		credential.Info.Balance += amount
		data, err = json.MarshalIndent(credential, "", " ")
		if err != nil {
			fmt.Println("Error at: Topup -> Error marshal credential")
			fmt.Println(err)
			return
		}
		err = os.WriteFile("./data/credential.json", data, 0644)
		if err != nil {
			fmt.Println("Error at: Topup -> Error update credetial")
			fmt.Println(err)
			return
		}
		//Print message
		fmt.Println("Balance update successfully!")
	}
}

func Withdraw() {
	//Check if client has been logged in
	data, err := os.ReadFile("./data/credential.json")
	if err != nil {
		fmt.Println("Error at: Withdraw -> Error reading crdential")
		fmt.Println(err)
		return
	}

	if len(data) == 0 {
		fmt.Println("You haven't logged in! This service required you to log in to continue")
		return
	}

	//Get token from credential
	var credential model.Credential
	err = json.Unmarshal(data, &credential)
	if err != nil {
		fmt.Println("Error at: Withdraw -> Error unmarshal crdential")
		fmt.Println(err)
		return
	}
	token := credential.Token

	var (
		amount  float64
		isValid bool
		reader  = bufio.NewReader(os.Stdin)
	)

	//Ask user's for amount of money
	isValid = false
	for !isValid {
		//Read amount from stdin
		fmt.Print("Enter your amount of withdraw: ")
		temp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: Withdraw -> Error reading amount from stdin")
			fmt.Println(err)
			return
		}
		temp = strings.TrimSpace(temp)
		amount, err = strconv.ParseFloat(temp, 64)
		if err != nil {
			fmt.Println("Invalid value for amount!")
			continue
		}

		//Check if amount is valid
		isValid = 0 < amount && amount <= credential.Info.Balance
		if !isValid {
			fmt.Printf("The amount to withdraw must be between 0 and %f\n", credential.Info.Balance)
		}
	}

	//Package data before sending
	data, err = json.MarshalIndent(amount, "", " ")
	if err != nil {
		fmt.Println("Error at: Withdraw -> Error marshal data before sending to server")
		fmt.Println(err)
		return
	}

	//Make new request
	url := "http://localhost:8800/withdraw"
	req, err := http.NewRequest("UPDATE", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error at: Withdraw -> Error making new request")
		fmt.Println(err)
		return
	}
	req.Header.Set("token", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at: Withdraw -> Error sending request to server or failed to receive respond")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	//Handele each respond status
	message, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error at: Withdraw -> Error reading respond body")
		fmt.Println(err)
		return
	}

	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error :(")
		return
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotAcceptable {
		fmt.Println(string(message))
		auth.Logout()
		return
	}

	if resp.StatusCode == http.StatusOK {
		//Update balance in credential
		credential.Info.Balance -= amount
		data, err = json.MarshalIndent(credential, "", " ")
		if err != nil {
			fmt.Println("Error at: Withdraw -> Error marshal credential")
			fmt.Println(err)
			return
		}
		err = os.WriteFile("./data/credential.json", data, 0644)
		if err != nil {
			fmt.Println("Error at: Withdraw > Error update credential")
			fmt.Println(err)
			return
		}
		//Print message to client
		fmt.Println("Balance updated successfully")
	}

}
