package main

import (
	"flag"
	"log"
	"os"

	"github.com/galogen13/yandex-go-metrics/internal/crypto"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	var (
		privateOut string
		publicOut  string
	)

	if err := logger.Initialize("info"); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Log.Sync()

	flag.StringVar(&privateOut, "private", "private.pem", "output file for private key")
	flag.StringVar(&publicOut, "public", "public.pem", "output file for public key")
	flag.Parse()

	logger.Log.Info("Generating RSA key pair")

	privatePEM, publicPEM, err := crypto.GenerateKeys()
	if err != nil {
		log.Fatalf("Failed to generate keys: %v\n", err)
	}

	if err := os.WriteFile(privateOut, []byte(privatePEM), 0600); err != nil {
		log.Fatalf("Failed to write private key: %v\n", err)
	}

	if err := os.WriteFile(publicOut, []byte(publicPEM), 0644); err != nil {
		log.Fatalf("Failed to write public key: %v\n", err)
	}

	logger.Log.Info("Keys generated successfully")
	logger.Log.Info("Private key", zap.String("path", privateOut))
	logger.Log.Info("Public key", zap.String("path", publicOut))
}
