package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/minio/minio-go/v7"
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

	if _, err := git.PlainClone(localCloneFolderPath, false, &git.CloneOptions{
		URL:      req.GithubRepoEndpoint,
		Progress: os.Stdout,
	}); err != nil {
		return err
	}

	defer os.RemoveAll(localCloneFolderPath)

	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	return filepath.WalkDir(localCloneFolderPath, func(path string, d fs.DirEntry, err error) error {

		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		absolutefilePath := cwd + "/" + strings.TrimPrefix(path, "./")
		key := strings.TrimPrefix(path, "local-clones/")

		fmt.Println("ABSOLUTE PATH: ", absolutefilePath)
		fmt.Println("OBJECT KEY: ", key)

		_, err = svc.app.MinioClient.FPutObject(context.Background(), svc.app.MinioBucketName, key, absolutefilePath, minio.PutObjectOptions{})

		return err
	})
}
