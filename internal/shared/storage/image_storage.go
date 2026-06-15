package storage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

// ─── Interface ─────────────────────────────────────────────────────────────────

// ImageStorage interface khusus untuk operasi file gambar
type ImageStorage interface {
	// UploadImage menyimpan file gambar dan mengembalikan URL publik
	UploadImage(file multipart.File, header *multipart.FileHeader, folder string) (string, error)

	// UploadImageWithThumbnail upload foto + buat thumbnail terkompresi
	// Mengembalikan (originalURL, thumbnailURL, error)
	UploadImageWithThumbnail(file multipart.File, header *multipart.FileHeader, folder string) (string, string, error)

	// DeleteImage menghapus file gambar berdasarkan URL publik
	DeleteImage(fileURL string) error

	// DeleteImageMultiple menghapus beberapa file gambar sekaligus
	DeleteImageMultiple(fileURLs ...string) []error
}

// ─── Config ────────────────────────────────────────────────────────────────────

// LocalImageConfig konfigurasi local image storage
type LocalImageConfig struct {
	BasePath   string // contoh: "./uploads"
	BaseURL    string // contoh: "http://localhost:1323/uploads"
	MaxSizeMB  int64  // default: 5
	ThumbnailW uint   // lebar thumbnail dalam pixel, default: 200
	ThumbnailH uint   // tinggi thumbnail (0 = maintain aspect ratio)
	Quality    int    // kualitas JPEG thumbnail (1-100), default: 75
}

// ─── Local Image Storage ───────────────────────────────────────────────────────

type localImageStorage struct {
	cfg LocalImageConfig
}

// NewLocalImageStorage membuat instance local image storage
func NewLocalImageStorage(cfg LocalImageConfig) ImageStorage {
	if cfg.MaxSizeMB == 0 {
		cfg.MaxSizeMB = 5
	}
	if cfg.ThumbnailW == 0 {
		cfg.ThumbnailW = 200
	}
	if cfg.Quality == 0 {
		cfg.Quality = 75
	}
	return &localImageStorage{cfg: cfg}
}

// UploadImage menyimpan file gambar ke local disk dan return URL publik
func (s *localImageStorage) UploadImage(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	// Validasi ukuran
	if header.Size > s.cfg.MaxSizeMB*1024*1024 {
		return "", fmt.Errorf("ukuran file melebihi batas %dMB", s.cfg.MaxSizeMB)
	}

	// Validasi ekstensi
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isAllowedImageExt(ext) {
		return "", fmt.Errorf("format file tidak didukung. Gunakan: jpg, jpeg, png, webp, heic")
	}

	// Buat direktori jika belum ada
	dir := filepath.Join(s.cfg.BasePath, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori: %w", err)
	}

	// Generate nama file gambar unik
	filename := generateImageFilename(ext)
	fullPath := filepath.Join(dir, filename)

	// Simpan file gambar
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file gambar: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("gagal menyimpan file gambar: %w", err)
	}

	// Return URL publik
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.cfg.BaseURL, "/"), folder, filename), nil
}

// UploadImageWithThumbnail upload foto asli + buat thumbnail terkompresi
func (s *localImageStorage) UploadImageWithThumbnail(file multipart.File, header *multipart.FileHeader, folder string) (string, string, error) {
	// Validasi ukuran
	if header.Size > s.cfg.MaxSizeMB*1024*1024 {
		return "", "", fmt.Errorf("ukuran file melebihi batas %dMB", s.cfg.MaxSizeMB)
	}

	// Validasi ekstensi
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isAllowedImageExt(ext) {
		return "", "", fmt.Errorf("format file tidak didukung. Gunakan: jpg, jpeg, png, webp")
	}

	// Baca konten file ke buffer agar bisa dipakai dua kali
	buf, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("gagal membaca file gambar: %w", err)
	}

	// Buat direktori
	dir := filepath.Join(s.cfg.BasePath, folder)
	thumbDir := filepath.Join(s.cfg.BasePath, folder, "thumbnails")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("gagal membuat direktori: %w", err)
	}
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return "", "", fmt.Errorf("gagal membuat direktori thumbnail: %w", err)
	}

	// Generate nama file unik (gunakan nama sama untuk original & thumbnail)
	baseName := generateImageFilename("")
	origFilename := baseName + ext
	thumbFilename := baseName + "_thumb.jpg" // thumbnail selalu JPEG

	// ─── Simpan foto original ──────────────────────────────────────────────────
	origPath := filepath.Join(dir, origFilename)
	if err := os.WriteFile(origPath, buf, 0644); err != nil {
		return "", "", fmt.Errorf("gagal menyimpan foto asli: %w", err)
	}

	// ─── Buat thumbnail terkompresi ────────────────────────────────────────────
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		// Gagal decode — hapus original yang sudah tersimpan
		os.Remove(origPath)
		return "", "", fmt.Errorf("gagal membaca gambar: %w", err)
	}

	// Resize: pertahankan aspect ratio
	thumbnail := resize.Resize(s.cfg.ThumbnailW, s.cfg.ThumbnailH, img, resize.Lanczos3)

	// Simpan thumbnail sebagai JPEG dengan kualitas tertentu
	thumbPath := filepath.Join(thumbDir, thumbFilename)
	thumbFile, err := os.Create(thumbPath)
	if err != nil {
		os.Remove(origPath)
		return "", "", fmt.Errorf("gagal membuat thumbnail: %w", err)
	}
	defer thumbFile.Close()

	if err := jpeg.Encode(thumbFile, thumbnail, &jpeg.Options{Quality: s.cfg.Quality}); err != nil {
		os.Remove(origPath)
		os.Remove(thumbPath)
		return "", "", fmt.Errorf("gagal mengkompresi thumbnail: %w", err)
	}

	baseURL := strings.TrimRight(s.cfg.BaseURL, "/")
	origURL := fmt.Sprintf("%s/%s/%s", baseURL, folder, origFilename)
	thumbURL := fmt.Sprintf("%s/%s/thumbnails/%s", baseURL, folder, thumbFilename)

	return origURL, thumbURL, nil
}

// DeleteImage menghapus file gambar dari local disk berdasarkan URL publik
func (s *localImageStorage) DeleteImage(fileURL string) error {
	if fileURL == "" {
		return nil
	}

	// Konversi URL ke path lokal
	localPath := s.imageUrlToPath(fileURL)
	if localPath == "" {
		return nil // URL tidak dikenali, skip
	}

	if err := os.Remove(localPath); err != nil {
		if os.IsNotExist(err) {
			return nil // File sudah tidak ada, tidak perlu error
		}
		return fmt.Errorf("gagal menghapus file gambar: %w", err)
	}

	return nil
}

// DeleteImageMultiple menghapus beberapa file gambar sekaligus
func (s *localImageStorage) DeleteImageMultiple(fileURLs ...string) []error {
	var errs []error
	for _, url := range fileURLs {
		if url == "" {
			continue
		}
		if err := s.DeleteImage(url); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

// imageUrlToPath mengkonversi URL publik ke path lokal
func (s *localImageStorage) imageUrlToPath(fileURL string) string {
	baseURL := strings.TrimRight(s.cfg.BaseURL, "/")
	if !strings.HasPrefix(fileURL, baseURL) {
		return ""
	}
	relativePath := strings.TrimPrefix(fileURL, baseURL)
	relativePath = strings.TrimPrefix(relativePath, "/")
	return filepath.Join(s.cfg.BasePath, relativePath)
}

// generateImageFilename membuat nama file gambar unik berdasarkan timestamp + random
func generateImageFilename(ext string) string {
	return fmt.Sprintf("img_%d_%d%s", time.Now().UnixNano(), time.Now().Unix()%1000, ext)
}

// isAllowedImageExt mengecek ekstensi file gambar yang diizinkan
func isAllowedImageExt(ext string) bool {
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".heic": true,
	}
	ext = strings.ToLower(ext)
	return allowed[ext]
}
