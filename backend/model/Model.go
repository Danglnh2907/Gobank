package model

import "time"

type User struct {
	ID string `json:"id"`
	//DateCreated time.Time `json:"date created"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Fullname string  `json:"fullname"`
	Balance  float64 `json:"balance"`
	Exp      int     `json:"exp"`
	State    string  `json:"state"`
}

type Admin struct {
	ID       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Info struct {
	ID       string  `json:"id"`
	Fullname string  `json:"fullname"`
	Role     string  `json:"role"`
	Balance  float64 `json:"balance"`
	Level    int     `json:"level"`
	Exp      int     `json:"exp"`
}

type Credential struct {
	Token string `json:"token"`
	Info  Info   `json:"info"`
}

type Transaction struct {
	Date          time.Time `json:"date transfer"`
	DebitAccount  string    `json:"debit account"`
	CreditAccount string    `json:"credit account"`
	Beneficiary   string    `json:"beneficiary"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
}
