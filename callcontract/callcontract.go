package contract

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	contract "github.com/NetSepio/dwifi-client/contract" // Assuming you have generated Go bindings for your contract
)

func MintAndPay(client *ethclient.Client, userAddress common.Address, privateKey *ecdsa.PrivateKey) (string, string, error) {
	// privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
	// if err != nil {
	// 	return fmt.Errorf("error loading private key: %v", err)
	// }

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", "", fmt.Errorf("error getting nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("error getting gas price: %v", err)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("error getting chain ID: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return "", "", fmt.Errorf("error creating transactor: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // Adjust as needed
	auth.GasPrice = gasPrice

	contractAddress := common.HexToAddress("0x8772540540639241C59Cc22e838FD8a0F2553EFf")
	instance, err := contract.NewContract(contractAddress, client)
	if err != nil {
		return "", "", fmt.Errorf("error creating contract instance: %v", err)
	}

	// Convert userAddress to string
	userAddressString := userAddress.Hex()

	// Call the mint function with the user's address as a string
	tx, err := instance.Mint(auth, userAddressString)
	if err != nil {
		return "", "", fmt.Errorf("error calling Mint function: %v", err)
	}

	fmt.Printf("Payment transaction sent: %s\n", tx.Hash().Hex())
	paymentTxHash := tx.Hash().Hex()

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return "", "", fmt.Errorf("error waiting for transaction to be mined: %v", err)
	}
	fmt.Printf("Payment transaction mined: %s\n", receipt.TxHash.Hex())
	minedTxHash := receipt.TxHash.Hex()

	return paymentTxHash, minedTxHash, nil
}
