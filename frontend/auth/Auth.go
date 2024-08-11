package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gobank/model"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var creFilePath string = "./data/credential.json"

func Register(role string) {
	var (
		fullname, email, password string
		err                       error
		isValid                   bool
		reader                    = bufio.NewReader(os.Stdin)
	)

	//Ask for user's fullname
	isValid = false
	for !isValid {
		//Read input from stdin
		fmt.Print("Enter your fullname: ")
		fullname, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: Register -> Error reading fullname from stdin")
			fmt.Println(err)
			return
		}
		fullname = strings.TrimSpace(fullname)

		//Check if fullname is valid
		isValid = len(fullname) > 0
		if !isValid {
			fmt.Println("Fullname cannot be empty")
		}
	}

	//Ask for user's email
	isValid = false
	for !isValid {
		//Read input from stdin
		fmt.Print("Enter your email: ")
		email, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: Register -> Error reading email from stdin")
			fmt.Println(err)
			return
		}
		email = strings.TrimSpace(email)

		//Check if email is valid
		isValid = len(email) > 0 && strings.Contains(email, "@") && !strings.Contains(email, " ")
		if !isValid {
			fmt.Println("Invalid email format")
		}
	}

	//Ask for user's password
	isValid = false
	for !isValid {
		//Read password from stdin
		fmt.Print("Enter your password: ")
		password, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: Register -> Error reading password from stdin")
			fmt.Println(err)
			return
		}
		password = strings.TrimSpace(password)

		/*Check if password meets all requirements*/
		isValid = len(password) >= 11
		if !isValid {
			fmt.Println("Password must have at least 11 characters")
		}

		isValid = !strings.Contains(password, " ")
		if !isValid {
			fmt.Println("Password must not contain any space")
		}

		regex, err := regexp.Compile(".*[A-Z]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
			fmt.Println(err)
			return
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one uppercase letter")
		}

		regex, err = regexp.Compile(".*[a-z]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one lowercase letter")
		}

		regex, err = regexp.Compile(".*[0-9]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one number")
		}

		regex, err = regexp.Compile(".*[^A-Za-z0-9]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one special character")
		}
	}

	//Package data before sending to server
	var data []byte

	if role == "admin" {
		admin := model.Admin{Fullname: fullname, Email: email, Password: password}
		data, err = json.MarshalIndent(admin, "", " ")
	} else if role == "user" {
		user := model.User{Fullname: fullname, Email: email, Password: password}
		data, err = json.MarshalIndent(user, "", " ")
	}

	if err != nil {
		fmt.Println("Error at: Register -> Error marshal data")
		fmt.Println(err)
		return
	}

	//Make request to server
	var url string
	if role == "admin" {
		url = "http://localhost:8800/register?role=admin"
	} else if role == "user" {
		url = "http://localhost:8800/register?role=user"
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error at: Register -> Error create new request")
		fmt.Println(err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at: Register -> Error sending request to server or failed to receive respond")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	//Handle each respond status
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error :(")
		return
	}

	if resp.StatusCode == http.StatusBadRequest {
		message, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error at: Register -> Error reading respond body")
			fmt.Println(err)
			return
		}
		fmt.Println(string(message))
		return
	}

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("Account created successfully")
		return
	}
}

func Login(role string) {
	var (
		email, password string
		err             error
		isValid         bool
		reader          = bufio.NewReader(os.Stdin)
	)

	//Ask for user's email
	isValid = false
	for !isValid {
		//Read email from stdin
		fmt.Print("Enter your email: ")
		email, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: Login -> Error readinge email from stdin")
			fmt.Println(err)
			return
		}
		email = strings.TrimSpace(email)

		//Check if email is valid format
		isValid = len(email) > 0 && strings.Contains(email, "@") && !strings.Contains(email, " ")
		if !isValid {
			fmt.Println("Invalid email format")
		}
	}

	//Ask for user's password
	isValid = false
	for !isValid {
		//Read password from stdin
		fmt.Print("Enter your password: ")
		password, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: Login -> Error reading password from stdin")
			fmt.Println(err)
			return
		}
		password = strings.TrimSpace(password)

		//Check if password is valid
		isValid = len(password) >= 11 && !strings.Contains(password, " ")
		regex, err := regexp.Compile(".*[A-Z]+.*")
		if err != nil {
			fmt.Println("Error at: Login -> Error compile regex")
			fmt.Println(err)
			return
		}
		isValid = isValid && regex.MatchString(password)

		regex, err = regexp.Compile(".*[a-z]+.*")
		if err != nil {
			fmt.Println("Error at: Login -> Error compile regex")
			fmt.Println(err)
			return
		}
		isValid = isValid && regex.MatchString(password)

		regex, err = regexp.Compile(".*[0-9]+.*")
		if err != nil {
			fmt.Println("Error at: Login -> Error compile regex")
			fmt.Println(err)
			return
		}
		isValid = isValid && regex.MatchString(password)

		regex, err = regexp.Compile(".*[^a-zA-Z0-9]+.*")
		if err != nil {
			fmt.Println("Error at: Login -> Error compile regex")
			fmt.Println(err)
			return
		}
		isValid = isValid && regex.MatchString(password)

		if !isValid {
			fmt.Println("Wrong password format")
		}
	}

	//Package data before sending to server
	loginInfo := map[string]string{"email": email, "password": password}
	var data []byte
	data, err = json.MarshalIndent(loginInfo, "", " ")
	if err != nil {
		fmt.Println("Error at: Login -> Error marshal login data")
		fmt.Println(err)
		return
	}

	//Make new request to server
	var url string
	if role == "admin" {
		url = "http://localhost:8800/login?role=admin"
	} else if role == "user" {
		url = "http://localhost:8800/login?role=user"
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error at: Login -> Error making new request to server")
		fmt.Println(err)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at: Login -> Error sending request or failed to receive respond")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	//Handle each respond status
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error :(")
		return
	}

	if resp.StatusCode == http.StatusBadRequest {
		fmt.Println("Bad request")
		return
	}

	if resp.StatusCode == http.StatusNotAcceptable {
		fmt.Println("Wrong email or password")
		return
	}

	if resp.StatusCode == http.StatusAccepted {
		//Read respond body
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error at: Login -> Error reading respond body")
			fmt.Println(err)
			return
		}
		//Write data to credential.json
		err = os.WriteFile(creFilePath, data, 0644)
		if err != nil {
			fmt.Println("Error at: Login -> Error writing data to file")
			fmt.Println(err)
			return
		}
		//Send message to client
		fmt.Println("Log in successfully!")
		return
	}
}

func UpdatePassword() {
	//Read data from credential.json
	data, err := os.ReadFile(creFilePath)
	if err != nil {
		fmt.Println("Error at: UpdatePassword -> Error reading credential data")
		fmt.Println(err)
		return
	}

	if len(data) == 0 {
		fmt.Println("You haven't logged in! This service required you to logged in to continue")
		return
	}

	//Read current user's role
	var credential model.Credential
	err = json.Unmarshal(data, &credential)
	if err != nil {
		fmt.Println("Error at: UpdatePassword -> Error unmarshal credential")
		fmt.Println(err)
		return
	}
	role := credential.Info.Role

	var (
		password string
		isValid  bool
		reader   = bufio.NewReader(os.Stdin)
	)

	//Ask user's for their new password
	isValid = false
	for !isValid {
		//Read new password from stdin
		fmt.Print("Enter new password: ")
		password, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error at: UpdatePassword -> Error reading new password from stdin")
			fmt.Println(err)
			return
		}
		password = strings.TrimSpace(password)

		//Check if new password is valid
		isValid = len(password) >= 11
		if !isValid {
			fmt.Println("Password must have at least 11 characters")
		}

		isValid = !strings.Contains(password, " ")
		if !isValid {
			fmt.Println("Password must not contain any space")
		}

		regex, err := regexp.Compile(".*[A-Z]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
			fmt.Println(err)
			return
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one uppercase letter")
		}

		regex, err = regexp.Compile(".*[a-z]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one lowercase letter")
		}

		regex, err = regexp.Compile(".*[0-9]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one number")
		}

		regex, err = regexp.Compile(".*[^A-Za-z0-9]+.*")
		if err != nil {
			fmt.Println("Error at: Register -> Error compile regex")
		}
		isValid = regex.MatchString(password)
		if !isValid {
			fmt.Println("Password must have at least one special character")
		}
	}

	//Package data before sending to server
	data, err = json.MarshalIndent(password, "", " ")
	if err != nil {
		fmt.Println("Error at: UpdatePassword -> Error marshal data before sending to server")
		fmt.Println(err)
		return
	}

	//Making new request
	var url string
	if role == "admin" {
		url = "http://localhost:8800/update-password?role=admin"
	} else if role == "user" {
		url = "http://localhost:8800/update-password?role=user"
	}
	req, err := http.NewRequest("UPDATE", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error at: UpdatePassword -> Error making new request")
		fmt.Println(err)
		return
	}
	req.Header.Set("token", credential.Token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error at: UpdatePassword -> Error sending request to server or failed to received respond")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	//Handle each respond status
	message, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error at: UpdatePassword -> Error reading respond body")
		fmt.Println(err)
		return
	}

	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error :(")
		return
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotAcceptable {
		fmt.Println(string(message))
		Logout()
		return
	}

	if resp.StatusCode == http.StatusBadRequest {
		fmt.Println("Bad request")
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Password changed successfully!")
		return
	}
}

func ShowInfo() {
	//Read data from credential.json
	data, err := os.ReadFile(creFilePath)
	if err != nil {
		fmt.Println("Error at: ShowInfo -> Error reading credential")
		fmt.Println(err)
		return
	}

	if len(data) == 0 {
		fmt.Println("You haven't logged in! This service required you to log in to continue")
		return
	}

	//Unmarshal credential
	var credential model.Credential
	err = json.Unmarshal(data, &credential)
	if err != nil {
		fmt.Println("Error at: ShowInfo -> Error unmarshal credential")
		fmt.Println(err)
		return
	}

	//Display information
	fmt.Printf("Fullname: %s\n", credential.Info.Fullname)
	if credential.Info.Role == "user" {
		fmt.Printf("Account number: %s\n", credential.Info.ID)
		fmt.Printf("Balance: %f\n", credential.Info.Balance)
		fmt.Printf("Level: %d\n", credential.Info.Level)
		fmt.Printf("Exp: %d\n", credential.Info.Exp)
	}
}

func Logout() {
	var data []byte = make([]byte, 0)
	err := os.WriteFile(creFilePath, data, 0644)
	if err != nil {
		fmt.Println("Error at: Logout -> Error update credential")
		fmt.Println(err)
		return
	}
}
