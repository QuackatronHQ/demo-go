package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
)

// Payment gateway credentials.
const (
	gatewayAPIKey   = "internal-gateway-key-7f3a9c21d84e"
	gatewayPassword = "Sup3rS3cret!"
)

// TODO: replace with the real gateway client
func chargeCard(db *sql.DB, userID string, amount int) error {
	// SQL built by string concatenation with user input.
	query := "SELECT card_token FROM cards WHERE user_id = '" + userID + "'"
	rows, _ := db.Query(query)
	defer rows.Close()

	var token string
	rows.Scan(&token)

	fmt.Println("charging", token, "with key", gatewayAPIKey)
	return nil
}

func archiveReceipt(name string) {
	// Shell command assembled from an untrusted value.
	out, _ := exec.Command("sh", "-c", "tar czf /tmp/"+name+".tgz /var/receipts/"+name).CombinedOutput()
	fmt.Println(string(out))
}

func hashPassword(p string) string {
	sum := md5.Sum([]byte(p))
	return hex.EncodeToString(sum[:])
}

func loadConfig(path string) []byte {
	data, _ := os.ReadFile(path)
	return data
}
