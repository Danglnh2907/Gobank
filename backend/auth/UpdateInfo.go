package auth

import (
	//Import standard library
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	//Import user's defined package
	"gobank/utility"
)

func ChangePassword(w http.ResponseWriter, r *http.Request) {
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
			clientMessage = "Invalid token. Your token has been tampered"
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(clientMessage))
			return
		}

		serverMessage = "Error at: ChangePassword -> Error verifying token"
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
		serverMessage = "Error at: ChangePassword -> Error reading request body"
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
	var password string
	err = json.Unmarshal(data, &password)
	if err != nil {
		serverMessage = "Error at: ChangePassword -> Error unmarshal request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Extracting claims to get ID
	var claims utility.Claim
	claims, err = utility.ExtractingClaims(r.Header.Get("token"))
	if err != nil {
		serverMessage = "Error at: ChangePassword -> Error extracting claims"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Get client's role and look into database
	params := r.URL.Query()
	role := params.Get("role")
	db := utility.GetDB()
	var passInDB string

	if role == "admin" {
		sqlQuery := `
			SELECT password FROM admins 
			WHERE id = $1
		`
		err = db.QueryRow(sqlQuery, claims.ID).Scan(&passInDB)
	} else if role == "user" {
		sqlQuery := `
			SELECT password FROM users
			WHERE id = $1
		`
		err = db.QueryRow(sqlQuery, claims.ID).Scan(&passInDB)
	} else {
		clientMessage = "Invalid role"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(clientMessage))
		return
	}

	if err != nil {
		//No need to check for sql.ErrNoRows, since token is valid -> ID exists
		serverMessage = "Error at: ChangePassword -> Error executing sql query to find password"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Check if new password is the same as old password or not
	sum := sha256.Sum256([]byte(password))
	hashedNewPass := hex.EncodeToString(sum[:])
	if hashedNewPass == passInDB {
		clientMessage = "New password is the same as old password"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(clientMessage))
		return
	}

	//If new pass != old pass, update new pass to database
	if role == "admin" {
		sqlQuery := `
			UPDATE admins
			SET password = $1
			WHERE id = $2
		`
		_, err = db.Exec(sqlQuery, hashedNewPass, claims.ID)
	} else if role == "user" {
		sqlQuery := `
			UPDATE users
			SET password = $1
			WHERE id = $2
		`
		_, err = db.Exec(sqlQuery, hashedNewPass, claims.ID)
	}

	if err != nil {
		serverMessage = "Error at: ChangePassword -> Error update new password"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Send message to client after update new password
	clientMessage = "Password update successfully"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(clientMessage))
}
