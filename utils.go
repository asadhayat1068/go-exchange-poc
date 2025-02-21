package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func transferETH(client *ethclient.Client, fromPriv *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()
	from, err := getAddress(fromPriv)
	if err != nil {
		return err
	}
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return err
	}
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(ctx)

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPriv)

	if err != nil {
		return err
	}

	return client.SendTransaction(ctx, signedTx)

}

func getAddress(privKey *ecdsa.PrivateKey) (common.Address, error) {
	pubKey := privKey.Public()
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*pubKeyECDSA)
	return address, nil
}
