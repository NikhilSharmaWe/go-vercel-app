package models

type RabbitMQResponse struct {
	ProjectID string `json:"projectID"`
	Service   string `json:"service"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}

type UploadRequest struct {
	GithubRepoEndpoint string `json:"githubRepoEndpoint"`
	ProjectID          string `json:"projectID"`
}
