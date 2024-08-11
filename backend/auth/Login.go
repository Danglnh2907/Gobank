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

	//Import user's defined package
	"gobank/model"
	"gobank/utility"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var clientMessage, serverMessage string

	//Read data from request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		serverMessage = "Error at: Login -> Error reading request body"
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
	var loginInfo map[string]string
	err = json.Unmarshal(data, &loginInfo)
	if err != nil {
		serverMessage = "Error at: Login -> Error unmarshal request body"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	/*Check for validity*/
	db := utility.GetDB()

	//Hash password for comparision
	sum := sha256.Sum256([]byte(loginInfo["password"]))
	password := hex.EncodeToString(sum[:])

	/*Check validity*/
	params := r.URL.Query()
	role := params.Get("role")

	var (
		admin model.Admin
		user  model.User
	)

	//Query to the respectful TABLE based on client's role
	if role == "admin" {
		sqlQuery := `
			SELECT * FROM admins
			WHERE email = $1 
		`
		err = db.QueryRow(sqlQuery, loginInfo["email"]).Scan(&admin.ID, &admin.Email, &admin.Password, &admin.Fullname)
	} else if role == "user" {
		sqlQuery := `
			SELECT * FROM users
			WHERE email = $1
		`
		err = db.QueryRow(sqlQuery, loginInfo["email"]).Scan(
			&user.ID, &user.Email, &user.Password, &user.Fullname, &user.Balance, &user.Exp, &user.State,
		)
	} else {
		clientMessage = "Invalid role"
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(clientMessage))
		return
	}

	if err != nil {
		//If not find the user, send messasage to client
		if err == sql.ErrNoRows {
			clientMessage = "This email has not been registered in the server"
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(clientMessage))
			return
		}
		/*Other error*/
		serverMessage = "Error at: Login -> Error querying database"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Compare password
	if (role == "admin" && admin.Password != password) || (role == "user" && user.Password != password) {
		clientMessage = "Wrong password"
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(clientMessage))
		return
	}

	/*If password match*/

	//Generate token
	var token string
	if role == "admin" {
		token, err = utility.GenerateToken(admin.ID, "admin")
	} else if role == "user" {
		token, err = utility.GenerateToken(user.ID, "user")
	}
	if err != nil {
		serverMessage = "Error at: Login -> Error generating token"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Generate credential struct and marshal to data
	var credential model.Credential
	if role == "user" {
		level := utility.CalculateLevel(user.Exp)
		credential = model.Credential{
			Token: token,
			Info: model.Info{
				ID:       user.ID,
				Fullname: user.Fullname,
				Role:     "user",
				Balance:  user.Balance,
				Level:    level,
				Exp:      user.Exp,
			},
		}
	} else if role == "admin" {
		credential = model.Credential{
			Token: token,
			Info: model.Info{
				ID:       admin.ID,
				Fullname: admin.Fullname,
				Role:     "admin",
			},
		}
	}
	data, err = json.MarshalIndent(credential, "", " ")
	if err != nil {
		serverMessage = "Error at: Login -> Error marshal data before sending to client"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Send data back to client
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(data))
}

func SendCredential(w http.ResponseWriter, r *http.Request) {
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
			clientMessage = "Cannot verify who you are! Your token may have been tampered"
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(clientMessage))
			return
		}

		/*Other errors*/
		serverMessage = "Error at: SendCredential -> Error verifying token"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Extracting claim
	claims, err := utility.ExtractingClaims(r.Header.Get("token"))
	if err != nil {
		serverMessage = "Error at: SendCredential -> Error extracting claims"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Get credential from database
	db := utility.GetDB()
	credential := model.Credential{
		Token: r.Header.Get("token"),
		Info: model.Info{
			ID:   claims.ID,
			Role: claims.Role,
		},
	}

	//Admin don't have any thing right now to update -> May change later
	if claims.Role == "user" {
		sqlQuery := `
			SELECT fullname, balance, exp FROM users
			WHERE id = $1
		`
		err = db.QueryRow(sqlQuery, claims.ID).Scan(&credential.Info.Fullname, &credential.Info.Balance, &credential.Info.Exp)
		if err != nil {
			serverMessage = "Error at: SendCredential -> Error executing sql query to find credential data"
			clientMessage = "Internal server error"

			//Log error to server
			fmt.Println(serverMessage)
			fmt.Println(err)

			//Send message to client
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(clientMessage))
			return
		}
	}

	//Calculate level
	credential.Info.Level = utility.CalculateLevel(credential.Info.Exp)

	//Package data
	data, err := json.MarshalIndent(credential, "", " ")
	if err != nil {
		serverMessage = "Error at: SendCredential -> Error marshal data"
		clientMessage = "Internal server error"

		//Log error to server
		fmt.Println(serverMessage)
		fmt.Println(err)

		//Send message to client
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(clientMessage))
		return
	}

	//Send data back to client
	w.WriteHeader(http.StatusFound)
	w.Write(data)
}
