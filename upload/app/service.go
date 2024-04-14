package app

import (
	"fmt"
	"os"
)

type UploadService interface {
	Upload(UploadRequest) error
}

type uploadService struct {
	app *Application
}

func NewUploadService(app *Application) UploadService {
	return &uploadService{
		app: app,
	}
}

func (svc *uploadService) Upload(req UploadRequest) error {
	localCloneFolderPath := fmt.Sprintf("./local-clones/%s", req.ProjectID)

	if err := clone(req.GithubRepoEndpoint, req.ProjectID); err != nil {
		return err
	}

	defer os.RemoveAll(localCloneFolderPath)

	return svc.app.pushToMinio(localCloneFolderPath)
}
