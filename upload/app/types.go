package app

type UploadRequest struct {
	GithubRepoEndpoint string `json:"githubRepoEndpoint"`
	ProjectID          string `json:"projectID"`
}

type UploadResponse struct {
	ProjectID string `json:"projectID"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}
