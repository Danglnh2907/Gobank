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

var creFilePath string = "./data/credential.json"

func MakeTransaction() {
	//Check if client has logged in
	data, err := os.ReadFile(creFilePath)
	if err != nil {
		fmt.Println("Error at: MakeTransaction -> Error reading credential")
		fmt.Println(err)
		return
	}

	if len(data) == 0 {
		fmt.Println("you haven't logged in! This service required you to log in to continue")
		return
	}

	//Get token from credential
	var credential model.Credential
	err = json.Unmarshal(data, &credential)
	if err != nil {
		fmt.Println("Error at: MakeTransaction -> Error unmarshal ccredential")
		fmt.Println(err)
		return
	}
	token := credential.Token

	var (
		transaction model.Transaction = model.Transaction{DebitAccount: credential.Info.ID}
		isValid     bool
		reader      = bufio.NewReader(os.Stdin)
	)

	/*Get transaction's data*/

	//Display source account information
	fmt.Println("Account information")
	fmt.Printf("\tDebit account: %s\n", credential.Info.ID)
	fmt.Printf("\tBalance: %f\n", credential.Info.Balance)
	fmt.Println(strings.Repeat("*", 20))

	//Ask for beneficiary's information
	isValid = false
	fmt.Println("Beneficiary information")
	for !isValid {
		//Read dest account number from stdin
		fmt.Print("Enter beneficiary account number: ")
		transaction.CreditAccount, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error reading destination account from stdin")
			fmt.Println(err)
			return
		}
		transaction.CreditAccount = strings.TrimSpace(transaction.CreditAccount)

		//Package credit account id to send to server
		data, err = json.MarshalIndent(transaction.CreditAccount, "", " ")
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error marhshal credit account")
			fmt.Println(err)
			return
		}

		//Find account from server
		url := "http://localhost:8800/fullname"
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error making new request")
			fmt.Println(err)
			return
		}
		req.Header.Set("token", token)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error sending request to server or failed to receive respond")
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		//Handle each respond status
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error reading respond body")
			fmt.Println(err)
			return
		}

		if resp.StatusCode == http.StatusInternalServerError {
			fmt.Println("Internal server error :(")
			return
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotAcceptable {
			fmt.Println(string(data))
			auth.Logout()
			return
		}

		if resp.StatusCode == http.StatusNotFound {
			fmt.Println("Cannot find any account with this ID")
			continue
		}

		if resp.StatusCode == http.StatusFound {
			//Display beneficiary's name
			transaction.Beneficiary = string(data)
			fmt.Printf("Beneficiary's name: %s\n", transaction.Beneficiary)
			fmt.Println(strings.Repeat("*", 20))
			isValid = true
		}
	}

	//Ask for transaction's amount
	fmt.Println("Transaction's information")
	isValid = false
	for !isValid {
		//Read amount from stdin
		fmt.Print("Enter amount of money: ")
		temp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error reading amount from stdin")
			fmt.Println(err)
			return
		}
		temp = strings.TrimSpace(temp)
		transaction.Amount, err = strconv.ParseFloat(temp, 64)
		if err != nil {
			fmt.Println("Invalid value for amount")
			continue
		}

		//Check if amount is a valid value
		isValid = 0 < transaction.Amount && transaction.Amount <= credential.Info.Balance
		if !isValid {
			fmt.Printf("Amount of money must be between 0 and %f\n", credential.Info.Balance)
		}
	}

	//Ask for transaction's description (no need for a loop since there's no edge cases)
	fmt.Print("Enter transaction's description (optional): ")
	transaction.Description, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error at: MakeTransaction -> Error reading description from stdin")
		fmt.Println(err)
		return
	}
	transaction.Description = strings.TrimSpace(transaction.Description)

	//Set default description
	if len(transaction.Description) == 0 {
		transaction.Description = fmt.Sprintf("%s transfer", credential.Info.Fullname)
	}

	//Ask user for confirmation
	fmt.Println(strings.Repeat("*", 20))
	fmt.Println("TRANSACTION'S DETAIL")
	fmt.Printf("\tDebit account: %s\n", transaction.DebitAccount)
	fmt.Printf("\tCredit account: %s\n", transaction.CreditAccount)
	fmt.Printf("\tBeneficiary's name: %s\n", transaction.Beneficiary)
	fmt.Printf("\tAmount: %f\n", transaction.Amount)
	fmt.Printf("\tDescription: %s\n", transaction.Description)

	//Get user's option
	isValid = false
	var option string
	for !isValid {
		fmt.Print("Confirmed? (Y/N) ")
		option, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at MakeTransaction -> Error reading user's option")
			fmt.Println(err)
			return
		}
		option = strings.ToUpper(strings.TrimSpace(option))

		isValid = option == "Y" || option == "N"
		if !isValid {
			continue
		}

		if option == "N" {
			return
		}

		if option == "Y" {
			isValid = true
		}
	}

	//Package data before sending to server
	data, err = json.MarshalIndent(transaction, "", " ")
	if err != nil {
		fmt.Println("Error at: MakeTransaction -> Error marshal transaction data")
		fmt.Println(err)
		return
	}

	//Make new request
	url := "http://localhost:8800/transaction"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error at: MakeTransaction -> Error making new request")
		fmt.Println(err)
		return
	}
	req.Header.Set("token", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at: MakeTransaction -> Error sending request to server or failed to receive respond")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	//Handle each respond status
	message, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error at: MakeTransaction -> Error reading request body")
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

	if resp.StatusCode == http.StatusCreated {
		//Update credential
		credential.Info.Balance -= transaction.Amount
		data, err = json.MarshalIndent(credential, "", " ")
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error marshal credential")
			fmt.Println(err)
			return
		}
		err = os.WriteFile(creFilePath, data, 0644)
		if err != nil {
			fmt.Println("Error at: MakeTransaction -> Error update crdential")
			fmt.Println(err)
			return
		}
		//Send message to client
		fmt.Println("Transaction created successfully")
		return
	}

}

func GetTransactions() {

}
