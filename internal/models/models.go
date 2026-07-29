package models

type Program struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	FolderPath  string   `json:"folderPath"`
	Tags        []string `json:"tags"`
}

type ScoredProgram struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	FolderPath  string  `json:"folderPath"`
	Similarity  float64 `json:"similarity"`
}

type SearchResult struct {
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Results []ScoredProgram `json:"results,omitempty"`
}

type LoadResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        string `json:"size"`
	Available   bool   `json:"available"`
	DownloadURL string `json:"download_url,omitempty"`
}
