package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("🚀 Starting DB Check...")
	db, err := sql.Open("sqlite", "trustflow.db")
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	fmt.Println("🔍 Checking Intents Table...")
	rows, err := db.Query("SELECT id, status, message FROM intents ORDER BY created_at DESC LIMIT 5")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, status, msg string
		if err := rows.Scan(&id, &status, &msg); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("📝 Intent: %s | Status: %s | Message: %s\n", id, status, msg)

		// Check steps for this intent
		stepRows, err := db.Query("SELECT step_index, action, status, tx_hash, error_msg FROM intent_steps WHERE intent_id = ? ORDER BY step_index", id)
		if err != nil {
			log.Fatal(err)
		}
		for stepRows.Next() {
			var idx int
			var action, sStatus, hash, errMsg sql.NullString
			if err := stepRows.Scan(&idx, &action, &sStatus, &hash, &errMsg); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("    ➡️ Step %d [%s]: %s (Tx: %s) Error: %s\n", idx, action.String, sStatus.String, hash.String, errMsg.String)
		}
		stepRows.Close()
		fmt.Println("---------------------------------------------------")
	}
}
