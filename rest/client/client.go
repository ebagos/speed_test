package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func uploadFile(filepath string) {
	// ファイルのオープン
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// HTTPリクエストの作成
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(part, f)
	if err != nil {
		log.Fatal(err)
	}
	contentType := writer.FormDataContentType()
	writer.Close()

	req, err := http.NewRequest("POST", "http://localhost:8080/upload", body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)

	// HTTPリクエストの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// レスポンスの処理
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
	}
	fmt.Println("Upload completed")
}

func downloadFile(filepath string) {
	// HTTPリクエストの作成
	req, err := http.NewRequest("GET", "http://localhost:8080/download?file="+filepath, nil)
	if err != nil {
		log.Fatal(err)
	}

	// HTTPリクエストの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// レスポンスの処理
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
	}

	// ファイルのオープン
	fileOut, err := os.Create(filepath + "_out")
	if err != nil {
		log.Fatal(err)
	}
	defer fileOut.Close()

	// レスポンスボディの書き込み
	_, err = io.Copy(fileOut, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Download completed")
}

func main() {
	// コマンドライン引数のパース
	var (
		file       string
		uploadFlag bool
	)
	flag.StringVar(&file, "file", "", "file to upload or download")
	flag.BoolVar(&uploadFlag, "upload", true, "true to upload, false to download")
	flag.Parse()

	// アップロードまたはダウンロードを実行
	if uploadFlag {
		uploadFile(file)
	} else {
		downloadFile(file)
	}
}
