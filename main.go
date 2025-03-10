package main

import (
	"chain-scan/db"
	"chain-scan/services"
	"log"
)

func main() {

	// Setup log file
	logFile, err := services.OpenFile("./log/eth_scan.log")
	if err != nil {
		log.Fatal("Could not set up log file.")
	}
	// Đảm bảo file được đóng khi chương trình kết thúc
	defer logFile.Close()

	// =============Websocket===============
	// =============Polygon=================
	// collectionPolygonWs := db.ConnectMongoDB("blockchainDB", "polygonWS")
	// services.HandleChainHTTP(collectionPolygonWs, "./configs/config-matic.json")
	// =============Arbitrum================
	// collectionArbitrumnWs := db.ConnectMongoDB("blockchainDB", "ArbitrumnWs")
	// services.HandleChainHTTP(collectionArbitrumnWs, "./configs/config-abr.json")
	// =============AvalancheWs=============
	// collectionAvalancheWs := db.ConnectMongoDB("blockchainDB", "AvalancheWs")
	// services.HandleChainHTTP(collectionAvalancheWs, "./configs/config-avax.json")
	// ==============Ethereum================
	// collectionEthereumWs := db.ConnectMongoDB("blockchainDB", "EthereumWs")
	// services.HandleChainHTTP(collectionEthereumWs, "./configs/config-eth.json")

	// =============Websocket===============
	// =============Polygon===============
	collectionPolygonWs := db.ConnectMongoDB("blockchainDB2", "transactions2")
	services.HandleChainWS(collectionPolygonWs, "./configs/config-matic.json")
	// =============AvalancheWs===============
	collectionAvalancheWs := db.ConnectMongoDB("blockchainDB2", "transactions2")
	services.HandleChainWS(collectionAvalancheWs, "./configs/config-avax.json")
	// =============Arbitrum===============
	collectionArbitrumnWs := db.ConnectMongoDB("blockchainDB2", "transactions2")
	services.HandleChainWS(collectionArbitrumnWs, "./configs/config-abr.json")
	// =============Ethereum===============
	collectionEthereumWs := db.ConnectMongoDB("blockchainDB2", "transactions2")
	services.HandleChainWS(collectionEthereumWs, "./configs/config-eth.json")

	// =============Solana===============
	collectionSolana := db.ConnectMongoDB("blockchainDB2", "transactions2")
	services.HandleChainSolana(collectionSolana, "./configs/config-sol.json")

	db.DisconnectMongoDB()
}
