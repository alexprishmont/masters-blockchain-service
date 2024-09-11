package signature

import (
	"blockchain-processor/internal/services/store"
	"context"
	"fmt"
	blockchainv1 "github.com/alexprishmont/masters-protos/gen/go/blockchain-processor"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"google.golang.org/grpc"
	"log"
	"log/slog"
)

type serverAPI struct {
	blockchainv1.UnimplementedBlockchainProcessorServer
	log      *slog.Logger
	contract *store.Store
	client   *ethclient.Client
	auth     *bind.TransactOpts
}

func Register(
	gRPC *grpc.Server,
	log *slog.Logger,
	contract *store.Store,
	client *ethclient.Client,
	auth *bind.TransactOpts,
) {
	blockchainv1.RegisterBlockchainProcessorServer(gRPC, &serverAPI{
		log:      log,
		contract: contract,
		client:   client,
		auth:     auth,
	})
}

func (s *serverAPI) SaveSignature(
	ctx context.Context,
	request *blockchainv1.SaveRequest,
) (*blockchainv1.SaveResponse, error) {

	balance, err := s.client.BalanceAt(ctx, s.auth.From, nil)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}
	fmt.Printf("Sender Balance: %d\n", balance)

	session := &store.StoreSession{
		Contract: s.contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From:     s.auth.From,
			Signer:   s.auth.Signer,
			GasLimit: 300000,
		},
	}

	tx, err := session.SaveSignature(
		request.GetId(),
		request.GetSignature(),
	)

	if err != nil {
		return nil, err
	}

	txHash := tx.Hash().Hex()
	s.log.Info("Tx sent", slog.String("tx", txHash))

	etherscanURL := fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", txHash)
	s.log.Info("Track your transaction at", slog.String("tracker_url", etherscanURL))

	return &blockchainv1.SaveResponse{Success: true}, nil
}

func (s *serverAPI) GetSignature(
	ctx context.Context,
	request *blockchainv1.GetRequest,
) (*blockchainv1.GetResponse, error) {
	signature, err := s.contract.GetSignature(&bind.CallOpts{}, request.GetId())

	if err != nil {
		return nil, err
	}

	if signature == "" {
		return &blockchainv1.GetResponse{
			Signature: signature,
			Success:   false,
		}, nil
	}

	return &blockchainv1.GetResponse{
		Signature: signature,
		Success:   true,
	}, nil
}
