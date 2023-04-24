package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"time"

	SpeedTest "github.com/ebagos/speed_test/proto"
	"google.golang.org/grpc"
)

func main() {
	// Parse command line arguments
	action := flag.String("action", "", "Upload or Download")
	filePath := flag.String("file", "", "Path to file")
	flag.Parse()

	// Validate command line arguments
	if *action == "" || (*action != "Upload" && *action != "Download") {
		log.Fatalf("Invalid action: %v", *action)
	}
	if *filePath == "" {
		log.Fatalf("File path is required")
	}
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %v", *filePath)
	}

	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create gRPC client
	client := SpeedTest.NewSpeedTestClient(conn)

	switch *action {
	case "Upload":
		upload(*filePath, client)
	case "Download":
		download(*filePath, client)
	}
}

func upload(filePath string, client SpeedTest.SpeedTestClient) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Call Upload RPC
	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatalf("Failed to upload: %v", err)
	}

	// Send file contents
	buffer := make([]byte, 1024)
	var totalBytes int64
	for {
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		err = stream.Send(&SpeedTest.UploadRequest{Data: buffer[:bytesRead]})
		if err != nil {
			log.Fatalf("Failed to send: %v", err)
		}
		totalBytes += int64(bytesRead)
	}

	// Close stream and receive response
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}
	log.Printf("Uploaded %v bytes %v", totalBytes, res.Bytes)
}

func download(filePath string, client SpeedTest.SpeedTestClient) {
	// Call Download RPC
	req := &SpeedTest.DownloadRequest{}
	downloadStream, err := client.Download(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Download data and write to file
	var totalBytes int64
	startTime := time.Now()
	for {
		res, err := downloadStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive: %v", err)
			bytesWritten, err := file.Write(res.Data)
			if err != nil {
				log.Fatalf("Failed to write file: %v", err)
			}
			totalBytes += int64(bytesWritten)
		}

		// Print download statistics
		elapsedTime := time.Since(startTime)
		log.Printf("Downloaded %v bytes in %v seconds (%v Mbps)", totalBytes, elapsedTime.Seconds(), float64(totalBytes)*8/1e6/elapsedTime.Seconds())
	}
}
