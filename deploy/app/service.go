package app

type DeployService interface {
	Deploy(DeployRequest) error
}

type deployService struct {
}

func NewDeployService() DeployService {
	return &deployService{}
}

func (*deployService) Deploy(DeployRequest) error {
	return nil
}
