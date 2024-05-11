package main

import (
	"blockchain-processor/internal/grpc/signature"
	"blockchain-processor/internal/services/store"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"google.golang.org/grpc"
	"log/slog"
	"math/big"
	"net"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "production"
)

func main() {
	gRPCServer := grpc.NewServer()
	log := setupLogger("local")

	ethClient, err := ethclient.Dial(os.Getenv("NODE_URL"))
	if err != nil {
		log.Error("failed to connect to Ethereum client", slog.Any("err", err))
	}
	contract, err := store.NewStore(
		common.HexToAddress(os.Getenv("CONTRACT_ADDRESS")),
		ethClient,
	)
	if err != nil {
		log.Error("failed to load contract", slog.Any("err", err))
	}

	privateKey, err := crypto.HexToECDSA(os.Getenv("WALLET_ADDRESS"))
	if err != nil {
		log.Error("Failed to load private key", slog.Any("err", err))
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111)) // ChainID 4 for Rinkeby
	if err != nil {
		log.Error("Failed to create authorized transactor", slog.Any("err", err))
	}

	signature.Register(
		gRPCServer,
		log,
		contract,
		ethClient,
		auth,
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", 44046))
	if err != nil {
		log.Error("error listening", slog.Any("err", err))
		return
	}

	if err := gRPCServer.Serve(l); err != nil {
		log.Error("error serving", slog.Any("err", err))
		return
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
