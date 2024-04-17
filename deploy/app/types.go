package app

type DeployRequest struct {
	ProjectID string
}

type DeployResponse struct {
	ProjectID string `json:"projectID"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}
