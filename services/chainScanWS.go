package services

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	RPC                 string `json:"rpc"`
	WssRPC              string `json:"wssRpc"`
	ETHContractAddress  string `json:"ethContractAddress"`
	USDTContractAddress string `json:"usdtContractAddress"`
	USDCContractAddress string `json:"usdcContractAddress"`
	WrappedBTCAddress   string `json:"wrappedBTCAddress"`
	TransferSignature   string `json:"transferSignature"`
	Chain               string `json:"chain"`
	TimeNeedToBlock     int    `json:"timeNeedToBlock"`
}

var config Config
var blockMutex sync.Mutex
var lastProcessedBlock = big.NewInt(0)

func LoadConfig(filePath string) (*Config, error) {
	// Mở file config
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	configFile := flag.String("config", "config-eth.json", "Path to the configuration file")
	flag.Parse()
	LoadConfig(*configFile)

	// Khởi tạo biến để lưu cấu hình
	var config Config

	// Đọc và decode file JSON
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	// Normalize các địa chỉ contract thành chữ thường
	config.ETHContractAddress = strings.ToLower(config.ETHContractAddress)
	config.USDTContractAddress = strings.ToLower(config.USDTContractAddress)
	config.USDCContractAddress = strings.ToLower(config.USDCContractAddress)
	config.WrappedBTCAddress = strings.ToLower(config.WrappedBTCAddress)

	// Trả về cấu hình đã đọc
	return &config, nil
}

func HandleChainWS(collection *mongo.Collection, fileConfig string) {

	// Load config before starting anything
	loadedConfig, err := LoadConfig(fileConfig)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	config = *loadedConfig
	// Create Ethereum client
	client, err := ethclient.Dial(config.WssRPC)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get MongoDB collection

	// Initialize lastProcessedBlock
	decimalLastProcessedBlock, err := client.BlockNumber(ctx)
	if err != nil {
		log.Printf("Failed to fetch latest block number: %v", err)
		return
	}
	lastProcessedBlock = big.NewInt(int64(decimalLastProcessedBlock))

	// Initialize log subscription and missed data handling
	go subscribeAndHandleLogs(ctx, client, collection)
	go monitorAndFetchMissedLogs(client, collection)

	// Block until termination signal
	waitForTermination()
}

func subscribeAndHandleLogs(ctx context.Context, client *ethclient.Client, collection *mongo.Collection) {
	logs := make(chan types.Log) // Bidirectional channel for sending and receiving
	for {
		if err := subscribeToLogs(ctx, client, logs, collection); err != nil {
			log.Printf("Log subscription error: %v. Retrying...", err)
		}
	}
}

func subscribeToLogs(ctx context.Context, client *ethclient.Client, logs chan types.Log, collection *mongo.Collection) error {
	fmt.Println("subscribeToLogs")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			common.HexToAddress(config.USDTContractAddress),
			common.HexToAddress(config.USDCContractAddress),
		},
		Topics: [][]common.Hash{
			{common.HexToHash(config.TransferSignature)},
		},
	}

	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %v", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case vLog := <-logs: // Receive logs from the channel
			processLog(client, vLog, collection)
		case err := <-sub.Err():
			return fmt.Errorf("subscription error: %v", err)
		case <-ctx.Done():
			return nil
		}
	}
}

func processLog(client *ethclient.Client, vLog types.Log, collection *mongo.Collection) {

	fmt.Println("processLog")
	blockNumber := big.NewInt(int64(vLog.BlockNumber))
	if blockNumber.Cmp(lastProcessedBlock) > 0 {
		fmt.Println("processLog block")
		log.Printf("Processing log from block: %d", blockNumber)

		// Marshal log để in dưới dạng JSON nếu cần
		txJSON, err := json.MarshalIndent(vLog, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal log: %v", err)
		}
		_, err = collection.InsertOne(context.Background(), vLog)
		if err != nil {
			log.Printf("Failed to insert log into MongoDB: %v", err)
		}

		log.Printf("Log JSON: %s\n", string(txJSON))

		// Cập nhật giá trị lastProcessedBlock sau khi xử lý log
		blockMutex.Lock()
		lastProcessedBlock.Set(blockNumber)
		blockMutex.Unlock()
	}

	if blockNumber.Cmp(lastProcessedBlock) < 0 {
		fetchMissedLogs(client, collection)
		return
	}
}

func fetchMissedLogs(client *ethclient.Client, collection *mongo.Collection) {
	fmt.Println("fetchMissedLogs")
	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Printf("Failed to fetch latest block number: %v", err)
		return
	}

	blockMutex.Lock()
	fromBlock := new(big.Int).Add(lastProcessedBlock, big.NewInt(1))
	blockMutex.Unlock()

	toBlock := big.NewInt(int64(latestBlock))
	if fromBlock.Cmp(toBlock) >= 0 {
		fmt.Println("fetchMissedLogseeeeeeeeeeeeeeeeeeeeeeeeeeee")
		return
	}

	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{
			common.HexToAddress(config.USDTContractAddress),
			common.HexToAddress(config.USDCContractAddress),
		},
		Topics: [][]common.Hash{
			{common.HexToHash(config.TransferSignature)},
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Printf("Failed to fetch missed logs: %v", err)
		return
	}

	for _, vLog := range logs {
		processLog(client, vLog, collection)
	}
}

func monitorAndFetchMissedLogs(client *ethclient.Client, collection *mongo.Collection) {
	fmt.Println("monitorAndFetchMissedLogs")
	for {
		time.Sleep(time.Duration(config.TimeNeedToBlock) * time.Microsecond)
		fetchMissedLogs(client, collection)
	}
}

func waitForTermination() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	log.Println("Shutting down...")
}
