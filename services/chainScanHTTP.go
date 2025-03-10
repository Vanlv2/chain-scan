package services

import (
	"chain-scan/configs"
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"os"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/ethereum/go-ethereum/rpc"
)

var walletAddresses = map[string]bool{
	strings.ToLower("0xd2f09CCF5e5cCd53ade1FefADc10492Bf03D3430"): true,
	strings.ToLower("0x88283c774eac4ED0462D101476D2C7C15AbD0FB3"): true,
}

func OpenFile(filePath string) (*os.File, error) {
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
		return nil, err
	}

	// Set log output to file
	log.SetOutput(logFile)

	return logFile, nil
}

func HandleChainHTTP(collection *mongo.Collection, fileConfig string) {

	config, err := configs.LoadConfig(fileConfig)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
		return
	}

	client, err := rpc.Dial(config.RPC)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum RPC: %v", err)
	}
	defer client.Close()

	// Get the latest block number
	var latestBlockHex string
	err = client.Call(&latestBlockHex, "eth_blockNumber")
	if err != nil {
		log.Fatalf("Failed to get latest block number: %v", err)
	}
	latestBlock := new(big.Int)
	latestBlock.SetString(latestBlockHex[2:], 16)

	// Start from the latest block and move backwards
	blockNumber := new(big.Int).Set(latestBlock)

	for i := 0; i < 2; i++ {
		err := processBlock(client, blockNumber, collection)
		if err == nil {
			// Giảm blockNumber để lùi về quá khứ
			blockNumber = new(big.Int).Sub(blockNumber, big.NewInt(1))
			time.Sleep(200 * time.Millisecond)
			continue
		}
		time.Sleep(time.Duration(config.TimeNeedToBlock) * time.Millisecond)
	}
}

func processBlock(client *rpc.Client, blockNumber *big.Int, collection *mongo.Collection) error {
	var block map[string]interface{}

	blockHex := fmt.Sprintf("0x%x", blockNumber)

	err := client.Call(&block, "eth_getBlockByNumber", blockHex, true)
	if err != nil {
		log.Printf("Failed to fetch block %s: %v", blockHex, err)
		return err
	}

	log.Printf("Processing block: %s - %s", blockHex, blockNumber.String())

	if block["transactions"] == nil {
		return fmt.Errorf("not found")
	}
	transactions := block["transactions"].([]interface{})

	for _, tx := range transactions {
		_, err := collection.InsertOne(context.Background(), tx)
		if err != nil {
			log.Printf("Error inserting transaction: %v", err)
		}
		txMap := tx.(map[string]interface{})
		processTransaction(txMap)
	}
	return nil
}

func handleERC20Transfer(input string) (string, *big.Int) {
	decodedData := decodeTransferInput(input[10:])
	if decodedData == nil {
		return "", nil
	}

	toAddress := strings.ToLower(decodedData["_to"].(string))
	value := new(big.Int)
	value.SetString(decodedData["_value"].(string), 10)

	if !walletAddresses[toAddress] {
		return "", nil
	}

	return toAddress, value

}

func decodeTransferInput(data string) map[string]interface{} {
	if len(data) < 128 {
		return nil
	}

	to := "0x" + strings.ToLower(data[24:64])
	valueHex := "0x" + data[64:128]
	value := new(big.Int)
	value.SetString(valueHex[2:], 16)

	return map[string]interface{}{
		"_to":    to,
		"_value": value.String(),
	}
}

func receiveTransfer(txHash string, token string, toAddress string, value *big.Int) error {

	// valueDiv := new(big.Int) // Tạo một big.Int để lưu kết quả
	// valueDiv.Div(value, big.NewInt(1000000))

	// Chuyển đổi thành *big.Float để xử lý số thập phân
	dividendFloat := new(big.Float).SetInt(value)

	var divisorFloat *big.Float
	if token == "eth" {
		// chia cho 12 chu so
		divisorFloat = big.NewFloat(1000000000000)
	} else {
		// chia cho 6 chu so
		divisorFloat = big.NewFloat(1000000)
	}

	// Thực hiện phép chia
	valueDiv := new(big.Float).Quo(dividendFloat, divisorFloat)

	log.Printf("ETH tx hash: %s, token: %s, to: %s, value: %.2f", txHash, token, toAddress, valueDiv)

	return nil
}

func processTransaction(tx map[string]interface{}) {
	to, ok := tx["to"].(string)
	if !ok || to == "" {
		return // Skip contract creation transactions
	}

	txHash, ok := tx["hash"].(string)
	if !ok || txHash == "" {
		return // Skip contract creation transactions
	}

	toAddress := strings.ToLower(to)
	valueHex, _ := tx["value"].(string)
	value := new(big.Int)
	value.SetString(valueHex[2:], 16)

	input, _ := tx["input"].(string)

	log.Printf("Transaction hash: %s, To: %s, Value: %s", txHash, toAddress, value.String())

	if config.Chain == "ethereum" && input == "0x" && value.Cmp(big.NewInt(0)) > 0 {
		// Direct ETH transfer
		if !walletAddresses[toAddress] {
			return
		}
		receiveTransfer(txHash, "eth", toAddress, value)
	}

	token := ""
	if toAddress == config.USDTContractAddress {
		token = "USDT"
	} else if toAddress == config.USDCContractAddress {
		token = "USDC"
	} else if toAddress == config.WrappedBTCAddress {
		token = "WBTC"
	} else if config.Chain != "ethereum" && toAddress == config.ETHContractAddress {
		token = "ETH"
	}

	if token == "" {
		return
	}

	if len(input) >= 10 && input[:10] == config.TransferSignature {
		toAddress, value := handleERC20Transfer(input)

		if value == nil {
			return
		}
		receiveTransfer(txHash, token, toAddress, value)

	}

}
