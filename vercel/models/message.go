package models

type RabbitMQResponse struct {
	ProjectID string `json:"project_id"`
	Service   string `json:"service"`
	Success   bool   `json:"success"`
}

type UploadRequest struct {
	GithubRepoEndpoint string `json:"repo_endpoint"`
	ProjectID          string `json:"project_id"`
}
