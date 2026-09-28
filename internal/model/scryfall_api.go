package model

import "time"

type ScryfallBulkResponse struct {
	ID          string    `json:"id"`
	UpdatedAt   time.Time `json:"updated_at"`
	DownloadURI string    `json:"jsonl_download_uri"`
}

type BulkSyncState struct {
	ID        string    `bson:"id"`
	UpdatedAt time.Time `bson:"updated_at"`
}
