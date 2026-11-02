package backup

// Status records whether backup output is safe to restore.
type Status string

const (
	// StatusPending marks a backup still being written.
	StatusPending Status = "pending"
	// StatusComplete marks a backup whose bytes and checksum were finalized.
	StatusComplete Status = "complete"
	// StatusInterrupted marks a canceled or interrupted backup.
	StatusInterrupted Status = "interrupted"
	// StatusFailed marks a backup that did not finish successfully.
	StatusFailed Status = "failed"
)
