package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

func NewSnapClient() *snap.Client {
	var client snap.Client

	client.New(
		os.Getenv("MIDTRANS_SERVER_KEY"),
		midtrans.Sandbox,
	)

	return &client
}

func VerifySignature(
	orderID string,
	statusCode string,
	grossAmount string,
	signature string,
	serverKey string,
) bool {

	raw := orderID + statusCode + grossAmount + serverKey
	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])

	return expected == signature
}
