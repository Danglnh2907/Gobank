package auth

import (
	//Import standard library
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	//Import user's defined package
	"gobank/model"
	"gobank/utility"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var clientMessage, serverMessage string

	//Reading request
	data, err := io.ReadAll(r.Body)
	if err != nil {
		serverMessage = "Error at: Register -> Error reading request body"
		clientMessage = "Internal server error"

		//Log errror to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send error message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Handle request based on request params
	params := r.URL.Query()
	role := params.Get("role")
	db := utility.GetDB()

	if role == "user" {
		//Unmarshal request body
		var user model.User = model.User{Balance: 0, Exp: 0, State: "active"}
		err = json.Unmarshal(data, &user)
		if err != nil {
			serverMessage = "Error at: Register -> Error unmarshal request body"
			clientMessage = "Internal server error"

			//Log error to server
			fmt.Println(serverMessage)
			fmt.Println(err)

			//Send error message to client
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(clientMessage))
			return
		}

		//Check if email has been registered in database
		sqlQuery := `
			SELECT id FROM users
			WHERE email = $1
		`
		var id string
		err = db.QueryRow(sqlQuery, user.Email).Scan(&id)

		//Handle error when executing sql query
		if err != nil && err != sql.ErrNoRows {
			serverMessage = "Error at: Register -> Error finding account in users TABLE"
			clientMessage = "Internal server error"

			//Log error to server
			fmt.Println(serverMessage)
			fmt.Println(err)

			//Send message to client
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(clientMessage))
			return
		}

		//If find no account (email has not been registered)
		if err == sql.ErrNoRows {
			//Calculate user's ID
			sqlQuery = "SELECT COUNT(*) FROM users"
			var numberOfUSers int64
			err = db.QueryRow(sqlQuery).Scan(&numberOfUSers)
			if err != nil {
				serverMessage = "Error at: Register -> Error getting number of users in TABLE users"
				clientMessage = "Internal server error"

				//Log error to server
				fmt.Println(serverMessage)
				fmt.Println(err)

				//Send message to client
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(clientMessage))
				return
			}
			user.ID = strconv.FormatInt(100000000+numberOfUSers, 10)

			//Hash password before storing
			sum := sha256.Sum256([]byte(user.Password))
			user.Password = hex.EncodeToString(sum[:]) //Using string(sum[:]) would likely to lead to non UTF8 chacracters

			//Store data into users TABLE
			sqlQuery = `
			INSERT INTO users (id, email, password, fullname, balance, exp, state)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
			_, err = db.Exec(sqlQuery, user.ID, user.Email, user.Password, user.Fullname, user.Balance, user.Exp, user.State)
			if err != nil {
				serverMessage = "Error at: Register -> Error insert data to database"
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
			w.WriteHeader(http.StatusCreated)
			clientMessage = "Account created successfully"
			w.Write([]byte(clientMessage))
			return
		}

		//If find any row (err == nil), notify user that account existed
		clientMessage = "This email has been registered in the system"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(clientMessage))
		return
	}

	if role == "admin" {
		//Unmarshal request body
		var admin model.Admin
		err = json.Unmarshal(data, &admin)
		if err != nil {
			serverMessage = "Error at: Register -> Error unmarshal request body"
			clientMessage = "Internal server error"

			//Log error to server
			fmt.Println(serverMessage)
			fmt.Println(err)

			//Send message to client
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(clientMessage))
			return
		}

		//Check if email has been registered in the database
		sqlQuery := `
			SELECT * FROM admins
			WHERE email = $1
		`
		var id int
		err = db.QueryRow(sqlQuery, admin.Email).Scan(&id)

		//Handle other errors when executing sql query
		if err != nil && err != sql.ErrNoRows {
			serverMessage = "Error at: Register -> Error finding account from admins TABLE"
			clientMessage = "Internal server error"

			//Log error to server
			fmt.Println(serverMessage)
			fmt.Println(err)

			//Send message to client
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(clientMessage))
			return
		}

		//If find no account
		if err == sql.ErrNoRows {
			//Hash password before storing
			sum := sha256.Sum256([]byte(admin.Password))
			admin.Password = hex.EncodeToString(sum[:])

			//Add data to database
			sqlQuery = "INSERT INTO admins (email, password, fullname) VALUES ($1, $2, $3)" //id is serial (auto generated)
			_, err = db.Exec(sqlQuery, admin.Email, admin.Password, admin.Fullname)
			if err != nil {
				serverMessage = "Error at: Register -> Error insert data to database"
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
			w.WriteHeader(http.StatusCreated)
			clientMessage = "Account created successfully"
			w.Write([]byte(clientMessage))
			return
		}

		//If find any row (err == nil), notify user that account existed
		clientMessage = "This email has been registered in the system"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(clientMessage))
		return
	}

	//If param is not either admin or user
	clientMessage = "Unidentified role"
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(clientMessage))
}
