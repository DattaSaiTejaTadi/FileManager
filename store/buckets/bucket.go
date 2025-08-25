package buckets

import (
	"encoding/json"
	"fm/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type buckets struct {
	baseURL      string
	bucketName   string
	serviceToken string
	client       *http.Client
}

func New(baseURL, bucketName, serviceToken string) *buckets {
	return &buckets{
		baseURL:      baseURL,
		bucketName:   bucketName,
		serviceToken: serviceToken,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (b *buckets) CreateFolder(fullPath string) (*models.CreateObjectResponse, *httperrors.Error) {
	url := fmt.Sprintf("%s/object/%s%s/.keep", b.baseURL, b.bucketName, fullPath)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, httperrors.NewDBError()
	}

	req.Header.Set("Authorization", b.serviceToken)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, httperrors.NewDBError()
	}

	defer resp.Body.Close()

	var objectResponse models.CreateObjectResponse

	body, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &objectResponse)
	if err != nil {
		return nil, httperrors.NewDBError()
	}
	log.Print(objectResponse)
	return &objectResponse, nil
}

func (b *buckets) GeneratePresignedUploadURL(fullPath string) (*models.UploadSignedURLResponse, *httperrors.Error) {
	url := fmt.Sprintf("%s/object/upload/sign/%s%s", b.baseURL, b.bucketName, fullPath)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, httperrors.NewDBError()
	}

	req.Header.Set("Authorization", b.serviceToken)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, httperrors.NewDBError()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, httperrors.NewDBError()
	}

	var presignedURLResponse models.UploadSignedURLResponse
	body, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &presignedURLResponse)
	if err != nil {
		return nil, httperrors.NewDBError()
	}
	presignedURLResponse.URL = b.baseURL + presignedURLResponse.URL
	presignedURLResponse.S3Key = b.bucketName + fullPath
	log.Print(presignedURLResponse)
	return &presignedURLResponse, nil
}
