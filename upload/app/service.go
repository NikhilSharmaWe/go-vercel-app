package app

type UploadService interface {
	Upload(UploadRequest) error
}

type uploadService struct {
}

func NewProductService() UploadService {
	return &uploadService{}
}

func (*uploadService) Upload(UploadRequest) error {
	return nil
}
