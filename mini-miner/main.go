package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Block struct {
	Nonce any     `json:"nonce"`
	Data  [][]any `json:"data"`
}

func (b *Block) toMap() map[string]any {
	return map[string]any{"data": b.Data, "nonce": b.Nonce}
}

type APIResponse struct {
	Difficulty int   `json:"difficulty"`
	Block      Block `json:"block"`
}

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		return
	}
}

const (
	MIN = 1
	MAX = 10_000
)

func main() {
	input := getProblem()
	_, nonce := solve(input.Difficulty, input.Block.toMap())

	// create response body
	ans := map[string]int{"nonce": nonce}
	postBody, err := json.Marshal(ans)
	if err != nil {
		log.Fatal(err)
		return
	}
	res := postSolution(postBody)
	fmt.Println(string(res))

}

// todo optimize this fn
func assertDifficulty(v []byte, diff int) bool {
	var binaryStr string
	for _, b := range v {
		binaryStr += fmt.Sprintf("%08b", b)
	}

	zeros := strings.Repeat("0", diff)
	return zeros == binaryStr[:diff]
}

// todo- impl done channel pattern
// todo- type x struct - make this better
// todo- general optimization and cleanup
func solve(difficulty int, data map[string]any) (bool, int) {
	type x struct {
		valid bool
		value int
	}
	ch := make(chan x)
	// done := make(chan struct{})
	max := int(math.Pow(2, float64(difficulty)))

	go func() {
		for i := MIN; i < max+1; i++ {
			// go func(done <-chan struct{}, ch chan<- x, nonce int) {
			go func(c chan<- x, nonce int) {

				copiedMap := copyMap(data)
				copiedMap["nonce"] = nonce

				mapByte, err := json.Marshal(copiedMap)
				if err != nil {
					log.Fatal(err)
				}

				sum := sha256.Sum256(mapByte)

				c <- x{assertDifficulty(sum[:], difficulty), nonce}
			}(ch, i)
		}
	}()

	for val := range ch {
		if val.valid { // milgya
			return true, val.value
			// done <- struct{}{}
		}
	}

	return false, -1
}

func copyMap(data map[string]any) map[string]any {
	copiedMap := make(map[string]any)
	for key, value := range data {
		copiedMap[key] = value
	}
	return copiedMap
}

func getProblem() *APIResponse {
	url := "https://hackattic.com/challenges/mini_miner/problem?access_token=" + os.Getenv("access_token")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Fatal(err)
		return nil
	}
	return &apiResp
}

func postSolution(data []byte) []byte {
	url := "https://hackattic.com/challenges/mini_miner/solve?playground=1&access_token=" + os.Getenv("access_token")
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
