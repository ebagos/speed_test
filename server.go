package main

import (
	"crypto/rand"
	"io"
	"log"
	"net"

	SpeedTest "github.com/ebagos/speed_test/proto"
	"google.golang.org/grpc"
)

type server struct{}

func (s *server) Upload(stream SpeedTest.SpeedTest_UploadServer) error {
	var totalBytes int64

	for {
		data, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&SpeedTest.UploadResponse{Bytes: totalBytes})
		}
		if err != nil {
			return err
		}
		totalBytes += int64(len(data.Data))
	}

	// return nil
}

func (s *server) Download(req *SpeedTest.DownloadRequest, stream SpeedTest.SpeedTest_DownloadServer) error {
	data := make([]byte, req.Bytes)
	_, err := io.ReadFull(rand.Reader, data)
	if err != nil {
		return err
	}

	for i := 0; i < len(data); i += 1024 {
		chunkSize := 1024
		if i+chunkSize > len(data) {
			chunkSize = len(data) - i
		}
		chunk := data[i : i+chunkSize]
		err := stream.Send(&SpeedTest.DownloadResponse{Data: chunk})
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	SpeedTest.RegisterSpeedTestServer(s, &server{})
	log.Printf("Server listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
