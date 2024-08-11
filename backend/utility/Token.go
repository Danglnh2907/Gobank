package utility

import (
	//Import standard library
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

var secretKey []byte = []byte("8DF72555912857A43FDAF8135B22A")

type Claim struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	IssueAt   time.Time `json:"issueAt"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func GenerateToken(id, role string) (string, error) {
	//Generate claim
	claim := Claim{
		ID:        id,
		Role:      role,
		IssueAt:   time.Now(),
		ExpiredAt: time.Now().Add(24 * time.Hour),
	}

	data, err := json.MarshalIndent(claim, "", " ")
	if err != nil {
		return "", err
	}

	//Genrate signature based on claims and secret key
	sum := sha256.Sum256(append(data, secretKey...))
	token := strings.Join([]string{hex.EncodeToString(data), hex.EncodeToString(sum[:])}, ".")
	return token, nil
}

type ExpiredTokenError struct{}

func (e ExpiredTokenError) Error() string {
	return "Token has expired"
}

type TokenTamperedError struct{}

func (e TokenTamperedError) Error() string {
	return "Error hashing"
}

func VerifyToken(token string) error {
	//Spliting the claims and signature
	tokenInfo := strings.Split(token, ".")

	//Decode the claims
	claimsData, err := hex.DecodeString(tokenInfo[0])
	if err != nil {
		return err
	}

	//Hash the claims with secret key and compare to the signature
	sum := sha256.Sum256(append(claimsData, secretKey...))
	if hex.EncodeToString(sum[:]) != tokenInfo[1] {
		return TokenTamperedError{}
	}

	//Check if the token is expired or not
	var claims Claim
	err = json.Unmarshal(claimsData, &claims)
	if err != nil {
		return err
	}

	if time.Now().After(claims.ExpiredAt) {
		return ExpiredTokenError{}
	}

	return nil
}

func ExtractingClaims(token string) (Claim, error) {
	//Spliting the claims and signature
	tokenInfo := strings.Split(token, ".")

	//Decode the claims
	claimsData, err := hex.DecodeString(tokenInfo[0])
	if err != nil {
		return Claim{}, err
	}

	//Unmarshal the claims
	var claims Claim
	err = json.Unmarshal(claimsData, &claims)
	if err != nil {
		return Claim{}, err
	}

	return claims, nil
}
