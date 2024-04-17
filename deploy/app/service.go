package app

import (
	"fmt"
	"os"
)

type DeployService interface {
	Deploy(DeployRequest) error
}

type deployService struct {
	app *Application
}

func NewDeployService(app *Application) DeployService {
	return &deployService{
		app: app,
	}
}

func (svc *deployService) Deploy(req DeployRequest) error {
	localCloneFolderPath := fmt.Sprintf("./local-clones/%s", req.ProjectID)

	objKeys, err := svc.app.getListOfAllFiles(req.ProjectID)
	if err != nil {
		return err
	}

	if err := svc.app.getFilesAndSaveLocally(objKeys); err != nil {
		return err
	}

	defer os.RemoveAll(localCloneFolderPath)

	if err := build(req.ProjectID); err != nil {
		return err
	}

	return svc.app.pushToMinio(localCloneFolderPath + "/dist")
}
