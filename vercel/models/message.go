package models

type RabbitMQResponse struct {
	ProjectID string `json:"projectID"`
	Type      string `json:"type"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}

type UploadRequest struct {
	GithubRepoEndpoint string `json:"githubRepoEndpoint"`
	ProjectID          string `json:"projectID"`
}

type ProjectRequest struct {
	GithubRepoEndpoint string `json:"githubRepoEndpoint"`
	ProjectID          string `json:"projectID"`
}

type DeployRequest struct {
	ProjectID string `json:"projectID"`
}
