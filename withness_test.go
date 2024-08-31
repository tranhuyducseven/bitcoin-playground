package main

import (
	"encoding/hex"
	"log"
	"testing"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/nghuyenthevinh2000/bitcoin-playground/testhelper"
	"github.com/stretchr/testify/assert"
)

// go test -v -run TestSendTxWithP2WPKH
func TestSendTxWithP2WPKH(t *testing.T) {
	suite := testhelper.TestSuite{}
	suite.SetupSimNetSuite(t, log.Default())

	// open from wallet
	fromWallet := suite.OpenWallet(t, DAVID_WALLET_SEED, "david2")
	fromAddress, err := fromWallet.CurrentAddress(0, waddrmgr.KeyScopeBIP0084)
	assert.Nil(t, err)
	fromPrivKey, err := fromWallet.PrivKeyForAddress(fromAddress)
	assert.Nil(t, err)
	fromPubKey, err := fromWallet.PubKeyForAddress(fromAddress)
	assert.Nil(t, err)
	fromName, err := fromWallet.AccountName(waddrmgr.KeyScopeBIP0084, 0)
	assert.Nil(t, err)
	compressedPubKey := fromPubKey.SerializeCompressed()

	hash160PubKey := hex.EncodeToString(btcutil.Hash160(compressedPubKey))

	_, err = suite.WalletClient.SendToAddress(fromAddress, btcutil.Amount(1e7))
	assert.Nil(t, err)

	// generate a block to confirm the transaction
	suite.GenerateBlocks(101)
	time.Sleep(15 * time.Second)

	utxos, err := fromWallet.ListUnspent(1, 9999999, fromName)
	assert.Nil(t, err)

	t.Log("Number of utxos: ", len(utxos))

	choosenUtxo := utxos[0]
	assert.Equal(t, choosenUtxo.ScriptPubKey[4:], hash160PubKey, "Error when fund new utxo")

	// open to wallet
	toWallet := suite.OpenWallet(t, BOB_WALLET_SEED, "bob2")
	toAddress, err := toWallet.CurrentAddress(0, waddrmgr.KeyScopeBIP0084)
	assert.Nil(t, err)

	fromBalance, err := fromWallet.CalculateBalance(1)
	assert.Nil(t, err)

	toBalance, err := toWallet.CalculateBalance(1)
	assert.Nil(t, err)

	t.Log("Before >> From balance: ", fromBalance)
	t.Log("Before >> To balance: ", toBalance)

	tx := wire.NewMsgTx(wire.TxVersion)

	// Add input
	utxoHash, err := chainhash.NewHashFromStr(choosenUtxo.TxID)
	assert.Nil(t, err)
	outPoint := wire.NewOutPoint(utxoHash, choosenUtxo.Vout)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)

	// Add output
	toAddressScript, err := txscript.PayToAddrScript(toAddress) // p2wpkh pubkey script
	assert.Nil(t, err)

	amount := int64(choosenUtxo.Amount*1e8 - 1e5)
	t.Log(">> Amount: ", amount)

	// add output
	txOut := wire.NewTxOut(amount, toAddressScript)
	tx.AddTxOut(txOut)

	pkScriptBytes, err := hex.DecodeString(choosenUtxo.ScriptPubKey)
	assert.Nil(t, err)

	// constructing witness field
	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		pkScriptBytes,
		int64(choosenUtxo.Amount),
	)

	// signing the transaction
	sigHashes := txscript.NewTxSigHashes(tx, inputFetcher)
	sig, err := txscript.
		RawTxInWitnessSignature(tx, sigHashes, 0, int64(choosenUtxo.Amount), pkScriptBytes, txscript.SigHashAll, fromPrivKey)

	assert.Nil(t, err)

	t.Log(">> Signature: ", hex.EncodeToString(sig))

	witness := wire.TxWitness{
		sig, compressedPubKey,
	}

	tx.TxIn[0].Witness = witness

	t.Logf(">> txIn[0]: %+v", tx.TxIn[0])

	rawCommitTx, err := suite.ChainClient.GetRawTransaction(utxoHash)
	assert.Nil(t, err)
	/* This scope is checked later */

	// check that this tx in is valid before sending
	blockUtxos := blockchain.NewUtxoViewpoint()
	sigCache := txscript.NewSigCache(50000)
	hashCache := txscript.NewHashCache(50000)

	blockUtxos.AddTxOut(btcutil.NewTx(rawCommitTx.MsgTx()), choosenUtxo.Vout, 1)
	hashCache.AddSigHashes(tx, inputFetcher)

	err = blockchain.ValidateTransactionScripts(
		btcutil.NewTx(tx), blockUtxos, txscript.ScriptVerifyWitnessPubKeyType, sigCache, hashCache,
	)
	assert.Nil(t, err)

	// send the raw transaction
	_, err = suite.WalletClient.SendRawTransaction(tx, false)
	assert.Nil(t, err)

	// generate a block to confirm the transaction
	time.Sleep(3 * time.Second)
	suite.GenerateBlocks(1)

	// check the balance of david
	time.Sleep(3 * time.Second)
	afterBalance, err := fromWallet.CalculateBalance(1)
	assert.Nil(t, err)
	t.Logf("After >> From balance: %d", afterBalance)

	// check the balance of bob
	afterBalance, err = toWallet.CalculateBalance(1)
	assert.Nil(t, err)
	t.Logf("After >> To balance: %d", afterBalance)

}
