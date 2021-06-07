package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/vulpemventures/go-elements/transaction"
	"io/ioutil"
	"net/http"
)
var (
	NotEnoughBalanceError = errors.New("Not enough balance on utxos")
)
type WalletStore interface {
	LoadPrivKey() (*btcec.PrivateKey, error)
	ListAddresses() ([]string, error)
}

type LiquiddWallet struct {
	Store WalletStore
}

func (s *LiquiddWallet) GetBalance() (uint64, error) {
	addresses, err := s.Store.ListAddresses()
	if err != nil {
		return 0, err
	}
	var balance uint64
	for _,v := range addresses {
		addressUnspents, err := unspents(v)
		if err != nil {
			return 0, err
		}
		for _,tx := range addressUnspents {
			balance += uint64(tx["value"].(float64))
		}
	}

	return balance, nil
}

func (s *LiquiddWallet) GetPubkey() (*btcec.PublicKey, error) {
	privkey, err := s.Store.LoadPrivKey()
	if err != nil {
		return nil, err
	}
	return privkey.PubKey(), nil
}

func (s *LiquiddWallet) GetPrivKey() (*btcec.PrivateKey, error) {
	return s.Store.LoadPrivKey()
}

// GetUtxos returns a slice of uxtos that match the given amount, as well as the change for the
func (s *LiquiddWallet) GetUtxos(amount uint64) ([]*transaction.TxInput, uint64, error) {
	addresses, err := s.Store.ListAddresses()
	if err != nil {
		return nil, 0, err
	}

	requiredBalance := amount
	var utxos []string
	for _,v := range addresses {
		addressUnspents, err := unspents(v)
		if err != nil {
			return nil,0, err
		}
		for _,tx := range addressUnspents {
			utxoValue := uint64(tx["value"].(float64))
			requiredBalance -= utxoValue
			utxos = append(utxos, tx[""])
		}
	}
	if  requiredBalance > 0 {
		return nil, 0, NotEnoughBalanceError
	}

}

func unspents(address string) ([]map[string]interface{}, error) {
	getUtxos := func(address string) ([]interface{}, error) {
		baseUrl, err := apiBaseUrl()
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("%s/address/%s/utxo", baseUrl, address)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s", data)
		var respBody interface{}
		if err := json.Unmarshal(data, &respBody); err != nil {
			return nil, err
		}
		return respBody.([]interface{}), nil
	}

	utxos := []map[string]interface{}{}
	for len(utxos) <= 0 {
		u, err := getUtxos(address)
		if err != nil {
			return nil, err
		}
		for _, unspent := range u {
			utxo := unspent.(map[string]interface{})
			utxos = append(utxos, utxo)
		}
	}

	return utxos, nil
}
func apiBaseUrl() (string, error) {
	return "http://localhost:3001", nil
}

type Transaction struct {

}