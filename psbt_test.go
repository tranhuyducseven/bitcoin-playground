package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/wire"
)

var txHex = "02000000000101b56d7a0fabbad41a7e885f96d6f30cf0c5484609b40caa8679eb99a47ccd52dd0000000000fdffffff04f82a0000000000002251206aa2f9333e2b1b4963adbb03a773017a5a852244bb3b69c7a200860da794f78c0000000000000000476a45010203040010aa19e27c19c156910cf863843850613fd282d30eec6a7127f7a37b0781790bec39f7c009e1af32243707ee226ea1d3002453836cad465753dc4d3d4c22d5e100000000000000003a6a380000000000000539f39fd6e51aad88f6f4ce6ab8827279cfffb92266faa7b3a4b5c3f54a934a2e33d34c7bc099f96cce0000000000002af86b0350090000000016001406fe8f3e467725e60d7c63edb07a1958f7a4f9b2024830450221009e1823d61880f22c5a6ff31ee117b20b85335d1e3399739c426ba89e39d4ff0d02203be263d109e412bf6c7b50b17a8e9aa367b48e54f050cfd2e1ce80aadbf7c30a01210210aa19e27c19c156910cf863843850613fd282d30eec6a7127f7a37b0781790b00000000"

func TestDecodeTx(t *testing.T) {

	txRaw, err := hex.DecodeString(txHex)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}

	// Create a new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Deserialize the transaction
	txReader := bytes.NewReader(txRaw)
	if err := tx.Deserialize(txReader); err != nil {
		t.Fatalf("Failed to deserialize transaction: %v", err)
	}

	// Print the transaction details
	fmt.Printf("Transaction: %+v\n", tx)
}
