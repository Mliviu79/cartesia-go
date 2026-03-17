package cartesia

import (
	"context"
	"io"
	"net/http"
)

// DatasetsFilesService manages files within datasets.
type DatasetsFilesService struct {
	client *Client
}

// DatasetFile represents a file in a dataset.
type DatasetFile struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
}

// DatasetFilesPage is a paginated list of dataset files.
type DatasetFilesPage = CursorPage[DatasetFile]

// FileUploadParams are the parameters for uploading a file to a dataset.
type FileUploadParams struct {
	File     io.Reader
	FileName string
	Purpose  string
}

// List returns files in a dataset.
func (s *DatasetsFilesService) List(ctx context.Context, datasetID string, params *ListParams) (*DatasetFilesPage, error) {
	var res DatasetFilesPage
	_, err := s.client.requestJSONQuery(ctx, http.MethodGet, "/datasets/"+datasetID+"/files", nil, params, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete removes a file from a dataset.
func (s *DatasetsFilesService) Delete(ctx context.Context, datasetID, fileID string) error {
	return s.client.requestDrain(ctx, http.MethodDelete, "/datasets/"+datasetID+"/files/"+fileID, nil)
}

// Upload uploads a file to a dataset using multipart form encoding.
func (s *DatasetsFilesService) Upload(ctx context.Context, datasetID string, params FileUploadParams) error {
	form := NewMultipartForm()

	if err := form.WriteFile("file", params.FileName, params.File); err != nil {
		return err
	}
	if params.Purpose != "" {
		if err := form.WriteField("purpose", params.Purpose); err != nil {
			return err
		}
	}
	if err := form.Close(); err != nil {
		return err
	}

	resp, err := s.client.requestMultipart(ctx, http.MethodPost, "/datasets/"+datasetID+"/files", form, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
