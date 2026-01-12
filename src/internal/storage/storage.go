package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"trustflow/src/pkg/types"

	_ "modernc.org/sqlite" // Import pure-Go sqlite driver
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	s := &Storage{db: db}
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

func (s *Storage) initSchema() error {
	createIntentsTable := `
	CREATE TABLE IF NOT EXISTS intents (
		id TEXT PRIMARY KEY,
		status TEXT,
		created_at INTEGER,
		message TEXT
	);`

	createStepsTable := `
	CREATE TABLE IF NOT EXISTS intent_steps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		intent_id TEXT,
		step_index INTEGER,
		action TEXT,
		tx_hash TEXT,
		status TEXT,
		error_msg TEXT,
		FOREIGN KEY(intent_id) REFERENCES intents(id)
	);`

	if _, err := s.db.Exec(createIntentsTable); err != nil {
		return err
	}
	if _, err := s.db.Exec(createStepsTable); err != nil {
		return err
	}
	return nil
}

func (s *Storage) SaveIntent(intent types.Intent) error {
	log.Printf("💾 Saving Intent: ID=%s", intent.ID)
	_, err := s.db.Exec("INSERT INTO intents (id, status, created_at) VALUES (?, ?, ?)",
		intent.ID, "pending", time.Now().Unix())
	if err != nil {
		log.Printf("❌ Failed to save intent %s: %v", intent.ID, err)
	} else {
		log.Printf("✅ Saved Intent %s", intent.ID)
	}
	return err
}

func (s *Storage) UpdateIntentStatus(id, status, message string) error {
	log.Printf("🔄 Updating Intent Status: ID=%s, Status=%s", id, status)
	_, err := s.db.Exec("UPDATE intents SET status = ?, message = ? WHERE id = ?", status, message, id)
	if err != nil {
		log.Printf("❌ Failed to update intent status %s: %v", id, err)
	}
	return err
}

func (s *Storage) SaveStep(intentID string, stepIndex int, action string) error {
	log.Printf("💾 Saving Step: IntentID=%s, Index=%d, Action=%s", intentID, stepIndex, action)
	_, err := s.db.Exec("INSERT INTO intent_steps (intent_id, step_index, action, status) VALUES (?, ?, ?, ?)",
		intentID, stepIndex, action, "pending")
	if err != nil {
		log.Printf("❌ Failed to save step for intent %s: %v", intentID, err)
	}
	return err
}

func (s *Storage) UpdateStepStatus(intentID string, stepIndex int, status, txHash, errorMsg string) error {
	log.Printf("🔄 Updating Step Status: IntentID=%s, Index=%d, Status=%s, TxHash=%s", intentID, stepIndex, status, txHash)
	_, err := s.db.Exec(`
		UPDATE intent_steps 
		SET status = ?, tx_hash = ?, error_msg = ? 
		WHERE intent_id = ? AND step_index = ?`,
		status, txHash, errorMsg, intentID, stepIndex)
	if err != nil {
		log.Printf("❌ Failed to update step status for intent %s: %v", intentID, err)
	}
	return err
}
