package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/crisgarner/go-ethereum/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	. "github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"os"
	"strconv"
)

var options  = []string{"Balance", "Mint", "Burn"}

func getEnv(key string) string{
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func main() {
	pk, err := crypto.HexToECDSA(getEnv("PRIVATE_KEY"))
	if err != nil{
		log.Fatalf("Could not load PK client: %v", err)
	}
	ethClient, err := client.New(context.Background(), &client.Config{
		URL: getEnv("CLIENT_URL"),
		Address: common.HexToAddress(getEnv("TOKEN_ADDRESS")),
		PrivateKey: *pk,
	})

	if err != nil {
		log.Fatalf("Could not initialize client: %v", err)
	}
	fmt.Println(Bold(Magenta("Token Tools")))
	menu(ethClient)
}

func menu(client *client.Broker){
	templates := promptui.SelectTemplates{
		Label: ` {{ "?" | cyan | bold }} {{ . | bold }}`,
		Active:   `{{ ">" | bold }} {{ . | cyan | bold }}`,
		Inactive: `  {{ . | cyan }}`,
		Selected: `  {{ "✔" | green | bold }} {{ "Select Option" | bold }}: {{ . | cyan }}`,
	}
	prompt := promptui.Select{
		Label: "Select Task",
		Items: options,
		Templates: &templates,
	}
	selected ,_, err := prompt.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch selected {
	case 0:
		account := Input("User Address",validAddress)
		balance, err := client.BalanceOf(common.HexToAddress(account))
		if err != nil{
			fmt.Println(err)
		}
		fmt.Println(Bold(Green(fmt.Sprintf("The Token Balance of account %s is: %s",account, balance.String()))))

	case 1:
		account := Input("User Address",validAddress	)
		amount := Input("Amount", validNumber)
		weiFloat, _ := strconv.ParseFloat(amount, 256)
		tx, err := client.Mint(common.HexToAddress(account), ToWei(weiFloat,18))
		if err != nil{
			fmt.Println(err)
		}
		fmt.Println(Bold(Green(fmt.Sprintf("Transaction submited with hash: %s", tx.Hash().Hex()))))
	case 2:
		account := Input("User Address",validAddress	)
		amount := Input("Amount", validNumber)
		weiFloat, _ := strconv.ParseFloat(amount, 256)
		tx, err := client.Burn(common.HexToAddress(account), ToWei(weiFloat,18))
		if err != nil{
			fmt.Println(err)
		}
		fmt.Println(Bold(Green(fmt.Sprintf("Transaction submited with hash: %s", tx.Hash().Hex()))))
	}
}

func Input(label string, validation func(input string) error)  string{
	prompt := promptui.Prompt{
		Label: label,
		Validate: validation,
	}
	keyword, err := prompt.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return keyword
}

func validAddress(input string) error {
	if !common.IsHexAddress(input) {
		return errors.New("address must be valid")
	}
	return nil
}

func validNumber(input string) error {
	_, err := strconv.ParseFloat(input, 256)
	if err != nil {
		return errors.New("amount must be numeric")
	}
	return nil
}


func ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}