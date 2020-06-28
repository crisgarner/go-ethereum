package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	//"math/big"
)

type Client interface {
	Config() *Config
}

type Config struct {
	URL 		string
	PrivateKey 	ecdsa.PrivateKey
	Address common.Address
}

type Broker struct {
	cfg    		*Config
	client 		*ethclient.Client
	contract	*Token
}

func New(ctx context.Context, cfg *Config) (*Broker, error) {
	if cfg.URL == "" {
		return nil, errors.New("must provide a INFURA URL value")
	}

	bkr := &Broker{
		cfg: cfg,
	}

	client, err := ethclient.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("could not initialize client: %v", err)
	}

	instance, err := NewToken(cfg.Address, client)

	if err != nil {
		log.Fatal(err)
	}

	bkr.client 		= client
	bkr.contract	= instance

	return bkr, nil
}

func (bkr *Broker) Config() *Config {
	return bkr.cfg
}

func (bkr *Broker) Client() *ethclient.Client {
	return bkr.client
}

func (bkr *Broker) BalanceOf(user common.Address) (*big.Int, error){
	balance,err := bkr.contract.BalanceOf(nil, user)
	if err != nil {
		return nil, err
	}
	return  balance, err
}

func (bkr *Broker) Mint(user common.Address, amount *big.Int) (*types.Transaction, error){
	auth := createTransaction(bkr)
	tx,err := bkr.contract.Mint(auth, user, amount)
	if err != nil {
		return nil, err
	}
	return  tx, err
}

func (bkr *Broker) Burn(user common.Address, amount *big.Int) (*types.Transaction, error){
	auth := createTransaction(bkr)
	tx,err := bkr.contract.Burn(auth, user, amount)
	if err != nil {
		return nil, err
	}
	return  tx, err
}


func createTransaction(bkr *Broker) *bind.TransactOpts {
	publicKey :=  bkr.cfg.PrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := bkr.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := bkr.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth 		   := bind.NewKeyedTransactor(&bkr.cfg.PrivateKey)
	auth.Nonce 		= big.NewInt(int64(nonce))
	auth.Value 		= big.NewInt(0)     // in wei
	auth.GasLimit 	= uint64(300000) // in units
	auth.GasPrice 	= gasPrice
	return auth
}
