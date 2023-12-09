package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

const sqlFile = "/tmp/decompressed.sql"

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
}

func cleanup() {
	os.Remove(sqlFile)
}

func main() {
	defer cleanup()

	// get problem input
	b64str := getProblem()

	// base64 decoding
	z9Compressed := b64Decode(b64str)

	// Z9 decompressing
	sql := gzipDecompress(z9Compressed)

	savetoFile(sql, sqlFile)

	// .sql file to postgres
	pgRestore(sqlFile)

	// get SSN as list
	ssn := fetchAliveSSN()

	// create response body
	jsonn := map[string][]string{"alive_ssns": ssn}
	postBody, err := json.Marshal(jsonn)
	if err != nil {
		log.Fatal(err)
		return
	}

	result := postSolution(postBody)
	log.Println("Result:", string(result))
}

func postSolution(data []byte) []byte {
	url := "https://hackattic.com/challenges/backup_restore/solve?access_token=" + os.Getenv("access_token")
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

func fetchAliveSSN() []string {
	q := "sudo -u postgres psql -d hackattic -c \"SELECT ssn FROM criminal_records WHERE status = 'alive'\""
	output, err := exec.Command("sh", "-c", q).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	// helper fn to clean output
	ssn := func(output []byte) []string {
		rawSSN := strings.Split(string(output), "-------------")[1]
		clutteredSSN := strings.Split(rawSSN, "\n")

		cleanSSN := []string{}
		for _, ssn := range clutteredSSN[:len(clutteredSSN)-3] {
			if ssn != "" {
				cleanSSN = append(cleanSSN, strings.TrimSpace(ssn))
			}
		}
		return cleanSSN
	}(output)

	return ssn
}

func pgRestore(filename string) {
	dropTableCMD := "sudo -u postgres psql -d hackattic -c \"DROP TABLE IF EXISTS public.criminal_records;\""
	_, err := exec.Command("sh", "-c", dropTableCMD).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	restoreCMD := "sudo -u postgres psql hackattic < " + filename
	_, err = exec.Command("sh", "-c", restoreCMD).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
}

func savetoFile(data []byte, filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	err = f.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

func b64Decode(b64str string) []byte {
	dec, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		log.Fatal(err)
	}
	return dec
}

func gzipDecompress(dec []byte) []byte {
	r, err := gzip.NewReader(bytes.NewReader(dec))
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	result, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func getProblem() string {
	url := "https://hackattic.com/challenges/backup_restore/problem?access_token=" + os.Getenv("access_token")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	defer resp.Body.Close()

	var apiResp struct {
		Dump string `json:"dump"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Fatal(err)
		return ""
	}

	return apiResp.Dump
}
