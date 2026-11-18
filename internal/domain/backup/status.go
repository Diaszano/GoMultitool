package backup

// Status records whether backup output is safe to restore.
type Status string

// StatusComplete marks a backup whose bytes and checksum were finalized.
const StatusComplete Status = "complete"
