package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

type APIResponse struct {
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Pbkdf2   struct {
		Rounds int    `json:"rounds"`
		Hash   string `json:"hash"`
	} `json:"pbkdf2"`
	Scrypt struct {
		N       int    `json:"N"`
		R       int    `json:"r"`
		P       int    `json:"p"`
		Buflen  int    `json:"buflen"`
		Control string `json:"_control"`
	} `json:"scrypt,omitempty"`
}

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	input := getProblem()

	decodedSalt := b64Decode(input.Salt)

	_sha256 := getSHA256([]byte(input.Password))
	_hmac := getHMAC(decodedSalt, []byte(input.Password))
	_pbkdf2 := getPBKDF2([]byte(input.Password), decodedSalt, input.Pbkdf2.Rounds, 32, sha256.New)
	_scrypt := getScrypt([]byte(input.Password), decodedSalt, input.Scrypt.N, input.Scrypt.R, input.Scrypt.P, input.Scrypt.Buflen)

	// create response body
	jsonn := map[string]string{
		"sha256": hex.EncodeToString(_sha256),
		"hmac":   hex.EncodeToString(_hmac),
		"pbkdf2": hex.EncodeToString(_pbkdf2),
		"scrypt": hex.EncodeToString(_scrypt),
	}
	postBody, err := json.Marshal(jsonn)
	if err != nil {
		log.Fatal(err)
		return
	}

	result := postSolution(postBody)
	fmt.Println("Result:", string(result))
}

func getScrypt(passwd, salt []byte, N, r, p, keyLen int) []byte {
	data, err := scrypt.Key(passwd, salt, N, r, p, keyLen)
	if err != nil {
		log.Fatal()
		return []byte{}
	}

	return data
}

func getPBKDF2(passwd, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	return pbkdf2.Key(passwd, salt, iter, keyLen, h)
}

func getHMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	output := h.Sum(nil)
	return output
}

func getSHA256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func b64Decode(b64str string) []byte {
	dec, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		log.Fatal(err)
	}
	return dec
}

func getProblem() APIResponse {
	url := "https://hackattic.com/challenges/password_hashing/problem?access_token=" + os.Getenv("access_token")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return APIResponse{}
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Fatal(err)
		return APIResponse{}
	}

	return apiResp
}

func postSolution(data []byte) []byte {
	url := "https://hackattic.com/challenges/password_hashing/solve?playground=1&access_token=" + os.Getenv("access_token")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
		return []byte{}
	}
	defer resp.Body.Close()

	res := make([]byte, resp.ContentLength)
	resp.Body.Read(res)
	return res
}
