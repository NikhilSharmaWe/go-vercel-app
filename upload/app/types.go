package app

type UploadRequest struct {
	GithubRepoEndpoint string `json:"repo_endpoint"`
	ProjectID          string `json:"project_id"`
}

type UploadResponse struct {
	ProjectID string `json:"project_id"`
	Success   bool   `json:"success"`
}
