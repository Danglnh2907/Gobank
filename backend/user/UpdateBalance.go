package user

import (
	"encoding/json"
	"fmt"
	"gobank/utility"
	"io"
	"net/http"
)

func Topup(w http.ResponseWriter, r *http.Request) {
	var serverMessage, clientMessage string

	//verify token
	err := utility.VerifyToken(r.Header.Get("token"))
	if err != nil {
		if _, ok := err.(utility.ExpiredTokenError); ok {
			clientMessage = "Your token has expired"
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
		serverMessage = "Error at: Topup -> Error verifying token"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Extracting claims
	var claims utility.Claim
	claims, err = utility.ExtractingClaims(r.Header.Get("token"))
	if err != nil {
		serverMessage = "Error at: Topup -> Error extracting claims"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Check if the requester has authority to perform this action
	if claims.Role == "admin" {
		clientMessage = "You have no authority to perfrom this action"
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(clientMessage))
		return
	}

	//Reading request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		serverMessage = "Error at: Topup -> Error reading request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}
	r.Body.Close()

	//Unmarshal request body
	var amount float64
	err = json.Unmarshal(data, &amount)
	if err != nil {
		serverMessage = "Error at: Topup -> Error unmarshal request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Query to database to update balance
	db := utility.GetDB()
	sqlQuery := `
		UPDATE users
		SET balance = balance + $1
		WHERE id = $2
	`
	_, err = db.Exec(sqlQuery, amount, claims.ID)
	if err != nil {
		serverMessage = "Error at: Topup -> Error executing sql query to update balance"
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
	clientMessage = "Balance update successfully"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(clientMessage))
}
