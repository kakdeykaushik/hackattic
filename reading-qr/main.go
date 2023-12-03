package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	imageURL := getProblem()
	imageBytes := downloadImage(imageURL)

	// create temp file
	file, err := os.CreateTemp("/tmp", "qr_*.png")
	if err != nil {
		log.Fatal(err)
	}
	// close and remove the file
	defer os.Remove(file.Name())
	defer file.Close()

	// write data to the file
	n, err := file.Write(imageBytes)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(n, "bytes written to the file", file.Name())

	// Get data out of QR code
	qrData := getQRData(file)

	// create response body
	qrJSON := map[string]string{"code": qrData}
	postBody, err := json.Marshal(qrJSON)
	if err != nil {
		log.Fatal(err)
		return
	}

	result := postSolution(postBody)
	fmt.Println("Result:", string(result))
}

func downloadImage(imageURL string) []byte {
	resp, err := http.Get(imageURL)
	if err != nil {
		log.Fatal(err)
		return []byte{}
	}
	defer resp.Body.Close()

	fmt.Println("Image Content Length", resp.ContentLength)
	imageBytes := make([]byte, resp.ContentLength)

	n, err := io.ReadFull(resp.Body, imageBytes)
	if err != nil {
		log.Fatal(err)
		return []byte{}
	}
	fmt.Println("Bytes read", n)
	return imageBytes[:n]
}

func getQRData(file *os.File) string {
	// after writing file pointer moved to EOF
	// move pointer back to 0 so that image can be read from beginning
	file.Seek(0, 0)

	img, format, err := image.Decode(file)
	fmt.Println("format: ", format)
	if err != nil {
		log.Fatal(err)
	}

	// prepare BinaryBitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		log.Fatal(err)
	}

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		log.Fatal(err)
	}

	return result.GetText()
}

func postSolution(data []byte) []byte {
	url := "https://hackattic.com/challenges/reading_qr/solve?playground=1&access_token=" + os.Getenv("access_token")
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

func getProblem() string {
	url := "https://hackattic.com/challenges/reading_qr/problem?access_token=" + os.Getenv("access_token")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	defer resp.Body.Close()

	var apiResp struct {
		ImageURL string `json:"image_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Fatal(err)
		return ""
	}

	return apiResp.ImageURL
}
