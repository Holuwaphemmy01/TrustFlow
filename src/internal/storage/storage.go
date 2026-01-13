package storage

import (
	"database/sql"
	"encoding/json"
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
		message TEXT,
		raw_intent TEXT
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

	// Migration for existing tables
	s.db.Exec("ALTER TABLE intents ADD COLUMN raw_intent TEXT")

	return nil
}

func (s *Storage) SaveIntent(intent types.Intent) error {
	log.Printf("💾 Saving Intent: ID=%s", intent.ID)
	rawBytes, _ := json.Marshal(intent)
	_, err := s.db.Exec("INSERT INTO intents (id, status, created_at, raw_intent) VALUES (?, ?, ?, ?)",
		intent.ID, "pending", time.Now().Unix(), string(rawBytes))
	if err != nil {
		log.Printf("❌ Failed to save intent %s: %v", intent.ID, err)
	} else {
		log.Printf("✅ Saved Intent %s", intent.ID)
	}
	return err
}

func (s *Storage) GetIntent(id string) (*types.IntentState, error) {
	// 1. Get Intent Details
	var state types.IntentState
	var rawIntent sql.NullString
	err := s.db.QueryRow("SELECT id, status, created_at, message, raw_intent FROM intents WHERE id = ?", id).
		Scan(&state.IntentID, &state.Status, &state.CreatedAt, &state.Message, &rawIntent)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch intent: %w", err)
	}
	if rawIntent.Valid {
		state.RawIntent = rawIntent.String
	}

	// 2. Get Steps
	rows, err := s.db.Query(`
		SELECT step_index, action, status, tx_hash, error_msg 
		FROM intent_steps 
		WHERE intent_id = ? 
		ORDER BY step_index ASC`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var step types.StepState
		var txHash, errorMsg sql.NullString // Handle nullable fields

		if err := rows.Scan(&step.StepIndex, &step.Action, &step.Status, &txHash, &errorMsg); err != nil {
			return nil, err
		}
		step.TxHash = txHash.String
		step.Error = errorMsg.String
		state.Steps = append(state.Steps, step)
	}

	return &state, nil
}

func (s *Storage) GetRecentIntents(limit int) ([]types.IntentState, error) {
	rows, err := s.db.Query("SELECT id, status, created_at, message FROM intents ORDER BY created_at DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch intents: %w", err)
	}
	defer rows.Close()

	var intents []types.IntentState
	for rows.Next() {
		var i types.IntentState
		if err := rows.Scan(&i.IntentID, &i.Status, &i.CreatedAt, &i.Message); err != nil {
			return nil, err
		}
		// We don't fetch steps here to keep listing lightweight
		intents = append(intents, i)
	}
	return intents, nil
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
