package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {

	base64Str := getProblem()

	unpacked := unpack(base64Str)

	postBody, err := json.Marshal(unpacked)
	if err != nil {
		log.Fatal(err)
		return
	}

	result := postSolution(postBody)
	fmt.Println("Result: ", string(result))
}

func unpack(enc []byte) map[string]any {
	dec := make([]byte, 32)
	base64.StdEncoding.Decode(dec, enc)

	// int signed - 4 bytes
	_int := int32(binary.LittleEndian.Uint32(dec[:4]))

	// uint - 4 bytes
	_uint := binary.LittleEndian.Uint32(dec[4:8])

	// short - 2bytes
	_short := int16(binary.LittleEndian.Uint16(dec[8:10]))

	// 2 bytes padding(from 10 to 12) - AA
	// float - 4 bytes
	_float := math.Float32frombits(binary.LittleEndian.Uint32(dec[12:16]))

	// _double - 8 bytes
	_double := math.Float64frombits(binary.LittleEndian.Uint64(dec[16:24]))

	// big endian double - 8 bytes
	_bigEndDouble := math.Float64frombits(binary.BigEndian.Uint64(dec[24:32]))

	return map[string]any{
		"int":               _int,          // 4 bytes
		"uint":              _uint,         // 4 bytes
		"short":             _short,        // 2 bytes
		"float":             _float,        // 4 bytes
		"double":            _double,       // 8 bytes
		"big_endian_double": _bigEndDouble, // 8 bytes
	}
}

func postSolution(data []byte) []byte {
	url := "https://hackattic.com/challenges/help_me_unpack/solve?playground=1&access_token=" + os.Getenv("access_token")
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

func getProblem() []byte {
	url := "https://hackattic.com/challenges/help_me_unpack/problem?access_token=" + os.Getenv("access_token")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return []byte{}
	}
	defer resp.Body.Close()

	var apiResp struct {
		Bytes string `json:"bytes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Fatal(err)
		return []byte{}
	}

	return []byte(apiResp.Bytes)
}
