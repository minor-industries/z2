package sync

type SyncResponse struct {
	ExistingItems int `json:"existing_items"`
	NewItems      int `json:"new_items"`
}
