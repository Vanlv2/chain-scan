package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/blocto/solana-go-sdk/client"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleChainSolana(collection *mongo.Collection, fileConfig string) {
	// Load config before starting anything
	loadedConfig, err := LoadConfig(fileConfig)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	config = *loadedConfig
	// Kết nối tới Solana node thông qua RPC endpoint
	c := client.NewClient(config.RPC)

	// Mở hoặc tạo file log
	file, err := os.OpenFile("transaction_solana.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open or create log file: %v", err)
	}
	defer file.Close()

	// Lấy thông tin slot mới nhất
	latestSlot, err := c.GetSlot(context.Background())
	if err != nil {
		log.Fatalf("Failed to get latest slot: %v", err)
	}

	// Ghi log thông tin slot mới nhất
	log.Printf("Latest slot: %d\n", latestSlot)

	// Lấy block dữ liệu từ slot mới nhất
	block, err := c.GetBlock(context.Background(), latestSlot)
	if err != nil {
		log.Fatalf("Failed to get block: %v", err)
	}
	// Ghi log giao dịch trong block
	logBlockTransactions(file, block, latestSlot, collection)
}

func logBlockTransactions(file *os.File, block *client.Block, slot uint64, collection *mongo.Collection) {
	// Ghi log thông tin về block
	logMessage := fmt.Sprintf("Block at slot #%d has %d transactions\n", slot, len(block.Transactions))
	fmt.Println(logMessage)
	if _, err := file.WriteString(logMessage); err != nil {
		log.Fatalf("Failed to write to log file: %v", err)
	}

	// Ghi log từng giao dịch trong block
	for _, tx := range block.Transactions {
		_, err := collection.InsertOne(context.Background(), tx)
		if err != nil {
			log.Printf("Error inserting transaction: %v", err)
		}
		// Tìm người gửi và người nhận từ danh sách accounts
		sender := tx.Transaction.Message.Accounts[0].ToBase58()
		receiver := tx.Transaction.Message.Accounts[1].ToBase58()

		// Tính toán giá trị giao dịch từ PreBalances và PostBalances
		preBalanceSender := tx.Meta.PreBalances[0]
		postBalanceSender := tx.Meta.PostBalances[0]
		transferredAmount := preBalanceSender - postBalanceSender // Giá trị được chuyển đi từ tài khoản người gửi

		// Ghi chi tiết người gửi, người nhận và giá trị giao dịch vào file log
		txDetails := fmt.Sprintf("Transaction:\nSender: %s\nReceiver: %s\nTransferred Amount: %d lamports\n", sender, receiver, transferredAmount)
		fmt.Println(txDetails)
		if _, err := file.WriteString(txDetails + "\n"); err != nil {
			log.Fatalf("Failed to write transaction details to log file: %v", err)
		}

		// Ghi log transaction dưới dạng JSON để kiểm tra thêm thông tin
		txJSON, err := json.MarshalIndent(tx, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal transaction: %v", err)
			continue
		}
		if _, err := file.WriteString(string(txJSON) + "\n"); err != nil {
			log.Fatalf("Failed to write transaction to log file: %v", err)
		}
	}
}
