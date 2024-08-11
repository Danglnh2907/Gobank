package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gobank/model"
	"gobank/utility"
	"io"
	"net/http"
)

func GetFullname(w http.ResponseWriter, r *http.Request) {
	var serverMessage, clientMessage string

	//Verify token
	err := utility.VerifyToken(r.Header.Get("token"))
	if err != nil {
		if _, ok := err.(utility.ExpiredTokenError); ok {
			clientMessage = "Your token has expired"
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(clientMessage))
			return
		}

		if _, ok := err.(utility.TokenTamperedError); ok {
			clientMessage = "We cannot verify who you are! Your token may have been tampered"
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(clientMessage))
			return
		}

		/*Other errors*/
		serverMessage = "Error at: FindAccount -> Error verifying token"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Reading request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		serverMessage = "Error at: FindAccount -> Error reading request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Unmarshal request body
	var id string
	err = json.Unmarshal(data, &id)
	if err != nil {
		serverMessage = "Error at: FindAccount -> Error unmarshal request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Querying into database to find the account's name
	db := utility.GetDB()
	sqlQuery := `
		SELECT fullname FROM users
		WHERE id = $1
	`
	var fullname string
	err = db.QueryRow(sqlQuery, id).Scan(&fullname)
	if err != nil {
		if err == sql.ErrNoRows {
			//Send message warning back to client
			clientMessage = "No account was found"
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(clientMessage))
			return
		}

		/*Other errors*/
		serverMessage = "Error at: FindAccount -> Error executing sql query to find account"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//If found, send data back to client
	data, err = json.MarshalIndent(fullname, "", " ")
	if err != nil {
		serverMessage = "Error at: FindAccount -> Error marshal data for sending to client"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(data)
	}

	w.WriteHeader(http.StatusFound)
	w.Write(data)
}

func MakeTransaction(w http.ResponseWriter, r *http.Request) {
	var serverMessage, clientMessage string

	//Verifying token
	err := utility.VerifyToken(r.Header.Get("token"))
	if err != nil {
		if _, ok := err.(utility.ExpiredTokenError); ok {
			clientMessage = "Your token has been expired"
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(clientMessage))
			return
		}

		if _, ok := err.(utility.TokenTamperedError); ok {
			clientMessage = "Cannot verify who you are! Your token has been tamper"
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(clientMessage))
			return
		}

		/*Other errors*/
		serverMessage = "Error at: MakeTransaction -> Error verifying token"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Extracting claims from token
	claims, err := utility.ExtractingClaims(r.Header.Get("token"))
	if err != nil {
		serverMessage = "Error at: MakeTransaction -> Error extracting claims"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Check if role is valid
	if claims.Role == "admin" {
		clientMessage = "You have no authority to perform this action"
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(clientMessage))
		return
	}

	//Read request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		serverMessage = "Error at: MakeTransaction -> Error reading request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Unmarshal request body
	var transaction model.Transaction
	err = json.Unmarshal(data, &transaction)
	if err != nil {
		serverMessage = "Error at: MakeTransaction -> Error unmarshal request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Add transaction to database
	db := utility.GetDB()
	sqlQuery := `
		INSERT INTO transactions (date, debit, credit, beneficiary, amount, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = db.Exec(sqlQuery,
		transaction.Date,
		transaction.DebitAccount,
		transaction.CreditAccount,
		transaction.Beneficiary,
		transaction.Amount,
		transaction.Description,
	)

	if err != nil {
		serverMessage = "Error at: MakeTransaction -> Error adding new transaction to database"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Update debit's balance
	sqlQuery = `
		UPDATE users
		SET balance = balance - $1
		WHERE id = $2
	`
	_, err = db.Exec(sqlQuery, transaction.Amount, transaction.DebitAccount)
	if err != nil {
		serverMessage = "Error at: MakeTransaction -> Error update debit's balance"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Update credit's balance
	sqlQuery = `
		UPDATE users
		SET balance = balance + $1
		WHERE id = $2
	`
	_, err = db.Exec(sqlQuery, transaction.Amount, transaction.CreditAccount)
	if err != nil {
		serverMessage = "Error at: MakeTransaction -> Error update credit's balance"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Send successful message to client
	clientMessage = "Transaction success"
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(clientMessage))
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	var serverMessage, clientMessage string

	//Verify token
	err := utility.VerifyToken(r.Header.Get("token"))
	if err != nil {
		if _, ok := err.(utility.ExpiredTokenError); ok {
			clientMessage = "Your token has been tampered"
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(clientMessage))
			return
		}

		if _, ok := err.(utility.TokenTamperedError); ok {
			clientMessage = "Cannot verify who you are! Your token may have been tampered"
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(clientMessage))
			return
		}

		/*Other errors*/
		serverMessage = "Error at: GetTransactions -> Error verifying token"
		clientMessage = "Internal server error"

		//Send message to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Extracting claims
	claims, err := utility.ExtractingClaims(r.Header.Get("token"))
	if err != nil {
		serverMessage = "Error at: GetTransactions -> Error extracting claims"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Check if role is valid
	if claims.Role == "admin" {
		clientMessage = "You have no authority to perform this action"
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(clientMessage))
		return
	}

	//Get transactions history from database
	//db := utility.GetDB()
	//sqlQuery := `
	//	SELECT TOP 20 * FROM transactions

	//`
}
