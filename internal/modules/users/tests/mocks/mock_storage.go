package mocks

import (
	"io"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

type MockImageStorage struct {
	mock.Mock
}

// ✅ PERBAIKAN: Gunakan testify/mock dan pastikan signature SAMA PERSIS dengan interface aslinya
func (m *MockImageStorage) UploadImage(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	args := m.Called(file, header, folder)

	// Cek nil untuk menghindari panic jika return value pertama adalah nil
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *MockImageStorage) UploadImageWithThumbnail(file multipart.File, header *multipart.FileHeader, folder string) (string, string, error) {
	args := m.Called(file, header, folder)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockImageStorage) DeleteImage(fileURL string) error {
	args := m.Called(fileURL)
	return args.Error(0)
}

func (m *MockImageStorage) DeleteImageMultiple(fileURLs ...string) []error {
	iArgs := make([]interface{}, len(fileURLs))
	for i, v := range fileURLs {
		iArgs[i] = v
	}
	args := m.Called(iArgs...)

	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]error)
}

// MockReader adalah io.Reader sederhana untuk test upload
type MockReader struct {
	mock.Mock
}

func (m *MockReader) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

// MockMultipartFile adalah mock multipart.File untuk test
type MockMultipartFile struct {
	mock.Mock
	io.Reader
}

func (m *MockMultipartFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	args := m.Called(p, off)
	return args.Int(0), args.Error(1)
}

func (m *MockMultipartFile) Seek(offset int64, whence int) (int64, error) {
	args := m.Called(offset, whence)
	return args.Get(0).(int64), args.Error(1)
}
