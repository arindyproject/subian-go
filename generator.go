package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

// ─── Config ────────────────────────────────────────────────────────────────────

type ModuleConfig struct {
	MainModule    string // contoh: artikel
	SubModule     string // contoh: kategori (atau artikel jika tidak ada -add)
	ModuleName    string // snake_case dari SubModule, untuk nama file & tabel
	ModuleTitle   string // PascalCase gabungan, contoh: ArtikelKategori atau Artikel
	ModulePlural  string // contoh: ArtikelKategoris
	PackageName   string // contoh: kategori
	ProjectModule string // contoh: neosim_go
	Timestamp     string // contoh: 20240507120000
	URLPrefix     string // contoh: /api/v1/artikel atau /api/v1/artikel/kategori
	TableName     string // contoh: artikels atau artikel_kategoris
}

// ─── Main ──────────────────────────────────────────────────────────────────────

func main() {
	name := flag.String("name", "", "Nama module utama (snake_case, contoh: artikel)")
	add := flag.String("add", "", "Nama sub-module (snake_case, contoh: kategori)")
	project := flag.String("project", "neosim_go", "Nama Go module project (dari go.mod)")
	flag.Parse()

	if *name == "" {
		log.Fatal("❌ Nama module utama wajib diisi. Contoh: go run ./cmd/gen/main.go -name=artikel")
	}

	subModule := *name
	if *add != "" {
		subModule = *add
	}

	mainPascal := toPascalCase(*name)
	subPascal := toPascalCase(subModule)

	entityTitle := mainPascal
	if *add != "" {
		entityTitle = mainPascal + subPascal
	}

	urlPrefix := fmt.Sprintf("/api/v1/%s", *name)
	if *add != "" {
		urlPrefix = fmt.Sprintf("/api/v1/%s/%s", *name, *add)
	}

	tableName := *name + "s"
	if *add != "" {
		tableName = *name + "_" + *add + "s"
	}

	cfg := ModuleConfig{
		MainModule:    *name,
		SubModule:     subModule,
		ModuleName:    subModule,
		ModuleTitle:   entityTitle,
		ModulePlural:  entityTitle + "s",
		PackageName:   toPackageName(subModule),
		ProjectModule: *project,
		Timestamp:     time.Now().Format("20060102150405"),
		URLPrefix:     urlPrefix,
		TableName:     tableName,
	}

	basePath := filepath.Join("internal", "modules", cfg.MainModule, cfg.SubModule)

	fmt.Printf("\n🚀 Membuat module: %s/%s\n", cfg.MainModule, cfg.SubModule)
	fmt.Printf("   Path: %s\n\n", basePath)

	files := buildFileList(cfg, basePath)
	for _, f := range files {
		if err := generateFile(f.path, f.tmpl, cfg); err != nil {
			log.Fatalf("❌ Gagal generate %s: %v", f.path, err)
		}
		fmt.Printf("   ✅ %s\n", f.path)
	}

	printNextSteps(cfg)
}

// ─── File List ─────────────────────────────────────────────────────────────────

type fileEntry struct {
	path string
	tmpl string
}

func buildFileList(cfg ModuleConfig, base string) []fileEntry {
	return []fileEntry{
		//contracts---------------------------------------------------------------
		{filepath.Join(base, "contracts", "interfaces.go"), tmplContracts},
		//dto---------------------------------------------------------------------
		{filepath.Join(base, "dto", fmt.Sprintf("%s_request.go", cfg.ModuleName)), tmplRequest},
		{filepath.Join(base, "dto", fmt.Sprintf("%s_response.go", cfg.ModuleName)), tmplResponse},
		//models------------------------------------------------------------------
		{filepath.Join(base, "models", fmt.Sprintf("%s.go", cfg.ModuleName)), tmplModel},
		//repositories------------------------------------------------------------
		{filepath.Join(base, "repositories", fmt.Sprintf("%s_repository.go", cfg.ModuleName)), tmplRepository},
		//services----------------------------------------------------------------
		{filepath.Join(base, "services", fmt.Sprintf("%s_service.go", cfg.ModuleName)), tmplService},
		//handlers----------------------------------------------------------------
		{filepath.Join(base, "handlers", fmt.Sprintf("%s_handler.go", cfg.ModuleName)), tmplHandler},
		//migrations--------------------------------------------------------------
		{filepath.Join(base, "migrations", fmt.Sprintf("%s_migrate.go", cfg.ModuleName)), tmplMigration},
		{filepath.Join(base, "migrations", fmt.Sprintf("001_create_%s_table.sql", cfg.TableName)), tmplSQL},
		//tests-------------------------------------------------------------------
		{filepath.Join(base, "tests", "factories", fmt.Sprintf("%s_factory.go", cfg.ModuleName)), tmplFactory},
		{filepath.Join(base, "tests", "seeders", fmt.Sprintf("%s_seeder.go", cfg.ModuleName)), tmplSeeder},
		{filepath.Join(base, "tests", "helpers", "db_helper.go"), tmplDBHelper},
		{filepath.Join(base, "tests", "mocks", fmt.Sprintf("%s_repository_mock.go", cfg.ModuleName)), tmplModuleServiceMock},
		{filepath.Join(base, "tests", "mocks", "rbac_repository_mock.go"), tmplRBACMock},
		{filepath.Join(base, "tests", "mocks", "auth_repository_mock.go"), tmplAuthMock},
		{filepath.Join(base, "tests", fmt.Sprintf("%s_service_test.go", cfg.ModuleName)), tmplModuleServiceTest},
		//main--------------------------------------------------------------------
		{filepath.Join(base, "module.go"), tmplModule},
		{filepath.Join(base, "routes.go"), tmplRoutes},
		{filepath.Join(base, "register.go"), tmplRegister},
	}
}

// ─── Generator ─────────────────────────────────────────────────────────────────

func generateFile(path, tmplStr string, cfg ModuleConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("   ⏭️  Skip (sudah ada): %s\n", path)
		return nil
	}

	tmpl, err := template.New(path).Parse(tmplStr)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, cfg)
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		result.WriteString(string(runes))
	}
	return result.String()
}

func toPackageName(s string) string {
	return strings.ReplaceAll(s, "_", "")
}

// ─── Next Steps ────────────────────────────────────────────────────────────────

func printNextSteps(cfg ModuleConfig) {
	fmt.Printf(`
────────────────────────────────────────────────────────
✅ Module '%s/%s' berhasil dibuat!

📋 Langkah selanjutnya:

1. Tambahkan blank import di internal/apps/apps.go:
   _ "%s/internal/modules/%s/%s"

2. Tambahkan blank import di cmd/migrate/main.go:
   _ "%s/internal/modules/%s/%s"

3. Edit model di:
   internal/modules/%s/%s/models/%s.go

4. Jalankan migrasi:
   make migrate-dev

5. Jalankan seeder:
   make seed
────────────────────────────────────────────────────────
`,
		cfg.MainModule, cfg.SubModule,
		cfg.ProjectModule, cfg.MainModule, cfg.SubModule,
		cfg.ProjectModule, cfg.MainModule, cfg.SubModule,
		cfg.MainModule, cfg.SubModule, cfg.ModuleName,
	)
}

// ─── Templates ─────────────────────────────────────────────────────────────────

// Contracts-----------------------------------------------------------
var tmplContracts = `package contracts

import (
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/dto"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
)

// AuthContext berisi informasi user yang sedang login untuk authorization
type AuthContext struct {
	UserID       int64
	IsSuperadmin bool
}

// Repository defines database operations
type Repository interface {
	Create(m *models.{{.ModuleTitle}}) error
	GetByID(id int64) (*models.{{.ModuleTitle}}, error)
	List(page, pageSize int, filter *dto.Filter{{.ModuleTitle}}Request) ([]models.{{.ModuleTitle}}, int64, error)
	Update(m *models.{{.ModuleTitle}}) error
	Delete(id int64) error
}

// Service defines business logic operations
type Service interface {
	Create(req *dto.Create{{.ModuleTitle}}Request, createdBy *int64, actor AuthContext) (*dto.{{.ModuleTitle}}Response, error)
	GetByID(id int64 , actor AuthContext) (*dto.{{.ModuleTitle}}Response, error)
	List(page, pageSize int,filter *dto.Filter{{.ModuleTitle}}Request, actor AuthContext) ([]dto.{{.ModuleTitle}}Response, int64, error)
	Update(id int64, req *dto.Update{{.ModuleTitle}}Request, updatedBy *int64, actor AuthContext) (*dto.{{.ModuleTitle}}Response, error)
	Delete(id int64, actor AuthContext) error
}
`

// Request-------------------------------------------------------------
var tmplRequest = `package dto

// Create{{.ModuleTitle}}Request request body untuk membuat {{.ModuleTitle}} baru
type Create{{.ModuleTitle}}Request struct {
	Name        string  ` + "`" + `json:"name" validate:"required,min=1,max=255"` + "`" + `
	Description *string ` + "`" + `json:"description" validate:"omitempty,max=500"` + "`" + `
}

// Update{{.ModuleTitle}}Request request body untuk update {{.ModuleTitle}}
type Update{{.ModuleTitle}}Request struct {
	Name        *string ` + "`" + `json:"name" validate:"omitempty,min=1,max=255"` + "`" + `
	Description *string ` + "`" + `json:"description" validate:"omitempty,max=500"` + "`" + `
}

// Filter{{.ModuleTitle}}Request request body untuk filter {{.ModuleTitle}}
type Filter{{.ModuleTitle}}Request struct {
	Name        string ` + "`" + `query:"name"` + "`" + `
}

`

// Response------------------------------------------------------------
var tmplResponse = `package dto

import (
	"time"

	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
)

// {{.ModuleTitle}}Response response untuk single {{.ModuleTitle}}
type {{.ModuleTitle}}Response struct {
	ID          int64     ` + "`" + `json:"id"` + "`" + `
	Name        string    ` + "`" + `json:"name"` + "`" + `
	Description *string   ` + "`" + `json:"description"` + "`" + `
	CreatedBy   *int64    ` + "`" + `json:"created_by"` + "`" + `
	UpdatedBy   *int64    ` + "`" + `json:"updated_by"` + "`" + `
	CreatedAt   time.Time ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt   time.Time ` + "`" + `json:"updated_at"` + "`" + `
}

// To{{.ModuleTitle}}Response mengubah model menjadi response
func To{{.ModuleTitle}}Response(m *models.{{.ModuleTitle}}) *{{.ModuleTitle}}Response {
	return &{{.ModuleTitle}}Response{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// To{{.ModuleTitle}}ListResponse mengubah slice model menjadi slice response
func To{{.ModuleTitle}}ListResponse(items []models.{{.ModuleTitle}}) []{{.ModuleTitle}}Response {
	var responses []{{.ModuleTitle}}Response
	for _, m := range items {
		responses = append(responses, *To{{.ModuleTitle}}Response(&m))
	}
	return responses
}
`

// Model---------------------------------------------------------------
var tmplModel = `package models

import (
	"time"

	"gorm.io/gorm"
)

// {{.ModuleTitle}} represents the {{.TableName}} table in database
type {{.ModuleTitle}} struct {
	ID          int64          ` + "`" + `gorm:"primaryKey;autoIncrement;column:id" json:"id"` + "`" + `
	Name        string         ` + "`" + `gorm:"column:name;type:varchar(255);not null" json:"name"` + "`" + `
	Description *string        ` + "`" + `gorm:"column:description;type:text" json:"description"` + "`" + `
	CreatedBy   *int64         ` + "`" + `gorm:"column:created_by" json:"created_by"` + "`" + `
	UpdatedBy   *int64         ` + "`" + `gorm:"column:updated_by" json:"updated_by"` + "`" + `
	CreatedAt   time.Time      ` + "`" + `gorm:"column:created_at;type:timestamptz;not null;default:NOW()" json:"created_at"` + "`" + `
	UpdatedAt   time.Time      ` + "`" + `gorm:"column:updated_at;type:timestamptz;not null;default:NOW()" json:"updated_at"` + "`" + `
	DeletedAt   gorm.DeletedAt ` + "`" + `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at"` + "`" + `
}

func ({{.ModuleTitle}}) TableName() string {
	return "{{.TableName}}"
}
`

// Repository----------------------------------------------------------
var tmplRepository = `package repositories

import (
	"errors"

	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/contracts"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/dto"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

// New{{.ModuleTitle}}Repository membuat instance repository baru
func New{{.ModuleTitle}}Repository(db *gorm.DB) contracts.Repository {
	return &repository{db: db}
}

func (r *repository) Create(m *models.{{.ModuleTitle}}) error {
	return r.db.Create(m).Error
}

func (r *repository) GetByID(id int64) (*models.{{.ModuleTitle}}, error) {
	var m models.{{.ModuleTitle}}
	result := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, result.Error
}

func (r *repository) List(page, pageSize int, filter *dto.Filter{{.ModuleTitle}}Request) ([]models.{{.ModuleTitle}}, int64, error) {
	var items []models.{{.ModuleTitle}}
	var total int64

	//------------------------------------------------------------
	// 1. Inisialisasi basis query & pastikan record yang di-soft delete tidak ikut terbawa
	query := r.db.Model(&models.{{.ModuleTitle}}{}).Where("deleted_at IS NULL")

	// 2. Filter Teks (Menggunakan ILIKE untuk case-insensitive)
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	
	// 3. Hitung total data berdasarkan filter yang aktif
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. Ambil data dengan paginasi dan sorting
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) Update(m *models.{{.ModuleTitle}}) error {
	return r.db.Save(m).Error
}

func (r *repository) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&models.{{.ModuleTitle}}{}).Error
}
`

// Service-------------------------------------------------------------
var tmplService = `package services

import (
	"errors"
	"time"
	"net/http"

	{{.ModuleName}}Contracts "{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/contracts"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/dto"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"

	//RBAC AUTH----------------------------------------
	authContracts "{{.ProjectModule}}/internal/modules/auth/contracts"
	rbacContracts "{{.ProjectModule}}/internal/modules/rbac/contracts"
	rbacMiddlewares "{{.ProjectModule}}/internal/modules/rbac/middlewares"
	rbacModels "{{.ProjectModule}}/internal/modules/rbac/models"
	appErrors "{{.ProjectModule}}/internal/shared/errors"
)

// ─── Init ───────────────────────────────────────────────────────────────────────
type service struct {
	repo {{.ModuleName}}Contracts.Repository
	rbacRepo rbacContracts.RBACRepository	//RBAC
	authRepo authContracts.AuthRepository	//AUTH
}

// New{{.ModuleTitle}}Service membuat instance service baru
func New{{.ModuleTitle}}Service(
	repo {{.ModuleName}}Contracts.Repository,
	rbacRepo rbacContracts.RBACRepository,	//RBAC
	authRepo authContracts.AuthRepository,	//AUTH
) {{.ModuleName}}Contracts.Service {
	return &service{
		repo: repo,
		rbacRepo: rbacRepo,	//RBAC
		authRepo: authRepo,	//AUTH
	}
}


// ─── Permission ─────────────────────────────────────────────────────────────────
func (s *service) canRead{{.ModuleTitle}}(actor {{.ModuleName}}Contracts.AuthContext) (bool, error) {
	if actor.IsSuperadmin {
		return true, nil
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyRead); err != nil || has {
		return has, err
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyManage); err != nil || has {
		return has, err
	}
	return false, nil
}

func (s *service) canCreate{{.ModuleTitle}}(actor {{.ModuleName}}Contracts.AuthContext) (bool, error) {
	if actor.IsSuperadmin {
		return true, nil
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyCreate); err != nil || has {
		return has, err
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyManage); err != nil || has {
		return has, err
	}
	return false, nil
}

func (s *service) canUpdate{{.ModuleTitle}}(actor {{.ModuleName}}Contracts.AuthContext) (bool, error) {
	if actor.IsSuperadmin {
		return true, nil
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyUpdate); err != nil || has {
		return has, err
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyManage); err != nil || has {
		return has, err
	}
	return false, nil
}

func (s *service) canDelete{{.ModuleTitle}}(actor {{.ModuleName}}Contracts.AuthContext) (bool, error) {
	if actor.IsSuperadmin {
		return true, nil
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyDelete); err != nil || has {
		return has, err
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermAnyManage); err != nil || has {
		return has, err
	}
	return false, nil
}

// ─── Service ────────────────────────────────────────────────────────────────────

// ------------ Create ------------------------------------------------------------
func (s *service) Create(req *dto.Create{{.ModuleTitle}}Request, createdBy *int64, actor {{.ModuleName}}Contracts.AuthContext) (*dto.{{.ModuleTitle}}Response, error) {
	// ─── Permission ─────────────────────────────────────────────────────────────
	can, err := s.canCreate{{.ModuleTitle}}(actor)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak memiliki hak akses untuk membuat {{.ModuleTitle}} baru.", nil)
	}

	// ─── Logic ──────────────────────────────────────────────────────────────────
	m := &models.{{.ModuleTitle}}{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return dto.To{{.ModuleTitle}}Response(m), nil
}

// ------------ GetByID -----------------------------------------------------------
func (s *service) GetByID(id int64, actor {{.ModuleName}}Contracts.AuthContext) (*dto.{{.ModuleTitle}}Response, error) {
	// ─── Permission ─────────────────────────────────────────────────────────────
	can, err := s.canRead{{.ModuleTitle}}(actor)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak memiliki hak akses untuk Melihat {{.ModuleTitle}}.", nil)
	}

	// ─── Logic ──────────────────────────────────────────────────────────────────
	m, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("{{.ModuleTitle}} tidak ditemukan")
	}
	return dto.To{{.ModuleTitle}}Response(m), nil
}

// ------------ List --------------------------------------------------------------
func (s *service) List(page, pageSize int,filter *dto.Filter{{.ModuleTitle}}Request, actor {{.ModuleName}}Contracts.AuthContext) ([]dto.{{.ModuleTitle}}Response, int64, error) {
	// ─── Permission ─────────────────────────────────────────────────────────────
	can, err := s.canRead{{.ModuleTitle}}(actor)
	if err != nil {
		return nil, 0, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, 0, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak memiliki hak akses untuk melihat daftar {{.ModuleTitle}}.", nil)
	}

	// ─── Logic ──────────────────────────────────────────────────────────────────
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	items, total, err := s.repo.List(page, pageSize, filter)
	if err != nil {
		return nil, 0, err
	}
	return dto.To{{.ModuleTitle}}ListResponse(items), total, nil
}

// ------------ Update ------------------------------------------------------------
func (s *service) Update(id int64, req *dto.Update{{.ModuleTitle}}Request, updatedBy *int64, actor {{.ModuleName}}Contracts.AuthContext) (*dto.{{.ModuleTitle}}Response, error) {
	// ─── Permission ─────────────────────────────────────────────────────────────
	can, err := s.canUpdate{{.ModuleTitle}}(actor)
	if err != nil {
		return nil, appErrors.Internal("gagal cek akses")
	}
	if !can {
		return nil, appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak memiliki hak akses untuk mengubah {{.ModuleTitle}}.", nil)
	}

	// ─── Logic ──────────────────────────────────────────────────────────────────
	m, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("{{.ModuleTitle}} tidak ditemukan")
	}
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Description != nil {
		m.Description = req.Description
	}
	m.UpdatedBy = updatedBy
	m.UpdatedAt = time.Now()

	if err := s.repo.Update(m); err != nil {
		return nil, err
	}
	return dto.To{{.ModuleTitle}}Response(m), nil
}

// ------------ Delete ------------------------------------------------------------
func (s *service) Delete(id int64, actor {{.ModuleName}}Contracts.AuthContext) error {
	// ─── Permission ─────────────────────────────────────────────────────────────
	can, err := s.canDelete{{.ModuleTitle}}(actor)
	if err != nil {
		return appErrors.Internal("gagal cek akses")
	}
	if !can {
		return appErrors.Wrap(http.StatusForbidden,
			"Akses ditolak. Anda tidak memiliki hak akses untuk menghapus {{.ModuleTitle}}.", nil)
	}

	// ─── Logic ──────────────────────────────────────────────────────────────────
	m, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if m == nil {
		return errors.New("{{.ModuleTitle}} tidak ditemukan")
	}
	return s.repo.Delete(id)
}
`

// Handler-------------------------------------------------------------
var tmplHandler = `package handlers

import (
	"net/http"
	"strconv"

	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/contracts"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/dto"
	"{{.ProjectModule}}/internal/shared/response"
	"{{.ProjectModule}}/internal/shared/validator"

	rbacMiddlewares "{{.ProjectModule}}/internal/modules/rbac/middlewares"

	"github.com/labstack/echo/v5"
)

// {{.ModuleTitle}}Handler defines HTTP handlers
type {{.ModuleTitle}}Handler struct {
	service contracts.Service
}

// New{{.ModuleTitle}}Handler membuat instance handler baru
func New{{.ModuleTitle}}Handler(service contracts.Service) *{{.ModuleTitle}}Handler {
	return &{{.ModuleTitle}}Handler{service: service}
}

// buildAuthContext membuat AuthContext dari JWT claims di context
func buildAuthContext(c *echo.Context) contracts.AuthContext {
	userID, _ := rbacMiddlewares.GetUserIDFromContext(c)
	isSuperadmin := rbacMiddlewares.IsSuperadmin(c)
	return contracts.AuthContext{
		UserID:       userID,
		IsSuperadmin: isSuperadmin,
	}
}

// ─── Private Helpers ───────────────────────────────────────────────────────────

func parseID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func getActorID(c *echo.Context) *int64 {
	if userID, ok := c.Get("userID").(int64); ok {
		return &userID
	}
	return nil
}

// ─── Handlers ──────────────────────────────────────────────────────────────────

// ─── List ──────────────────────────────────────────────────────────────────────
// {{.ModuleTitle}}Handler godoc
//
//	@Summary		Get list of {{.ModuleTitle}}
//	@Description	Get paginated list of {{.ModuleTitle}}
//	@Tags			{{.MainModule}}/{{.SubModule}}
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name			query		string	false	"Filter by name (partial match)"
//	@Param			page			query		int		false	"Page number"
//	@Param			page_size		query		int		false	"Page size"
//	@Success		200				{object}	response.MyGoResponse{data=[]dto.{{.ModuleTitle}}Response}
//	@Router			{{.URLPrefix}} [get]
//
// List handles GET {{.URLPrefix}}
func (h *{{.ModuleTitle}}Handler) List(c *echo.Context) error {
	page, pageSize := 1, 10

	// Mengambil query parameter untuk filter
	filter := dto.Filter{{.ModuleTitle}}Request{
		Name:     c.QueryParam("name"),
	}

	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	actor := buildAuthContext(c)
	items, total, err := h.service.List(page, pageSize,&filter, actor)
	if err != nil {
		return response.Response(c, http.StatusInternalServerError, false, "Gagal mengambil data", nil, nil)
	}

	return response.Paginated(c, http.StatusOK, true, "Berhasil mengambil data", items, total, page, pageSize)
}

// ─── GetByID ───────────────────────────────────────────────────────────────────
// {{.ModuleTitle}}Handler godoc
//
//	@Summary		Get {{.ModuleTitle}}
//	@Description	Get {{.ModuleTitle}} by :id
//	@Tags			{{.MainModule}}/{{.SubModule}}
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"{{.ModuleTitle}} ID"
//	@Success		200	{object}	response.MyGoResponse{data=dto.{{.ModuleTitle}}Response}
//	@Router			{{.URLPrefix}}/{id} [get]
//
// GetByID handles GET {{.URLPrefix}}/:id
func (h *{{.ModuleTitle}}Handler) GetByID(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	actor := buildAuthContext(c)
	item, err := h.service.GetByID(id, actor)
	if err != nil {
		return response.Response(c, http.StatusNotFound, false, err.Error(), nil, nil)
	}

	return response.Response(c, http.StatusOK, true, "Berhasil mengambil data", item, nil)
}

// ─── Create ────────────────────────────────────────────────────────────────────
// {{.ModuleTitle}}Handler godoc
//
//	@Summary		Create {{.ModuleTitle}}
//	@Description	Create New {{.ModuleTitle}}
//	@Tags			{{.MainModule}}/{{.SubModule}}
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.Create{{.ModuleTitle}}Request	true	"Create Request"
//	@Success		200		{object}	response.MyGoResponse{data=dto.{{.ModuleTitle}}Response}
//	@Router			{{.URLPrefix}} [post]
//
// Create handles POST {{.URLPrefix}}
func (h *{{.ModuleTitle}}Handler) Create(c *echo.Context) error {
	var req dto.Create{{.ModuleTitle}}Request
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}

	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	actor := buildAuthContext(c)
	item, err := h.service.Create(&req, getActorID(c), actor)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, err.Error(), nil, nil)
	}

	return response.Response(c, http.StatusCreated, true, "Data berhasil dibuat", item, nil)
}


// ─── Update ────────────────────────────────────────────────────────────────────
// {{.ModuleTitle}}Handler godoc
//
//	@Summary		Update {{.ModuleTitle}}
//	@Description	Update {{.ModuleTitle}} by :id
//	@Tags			{{.MainModule}}/{{.SubModule}}
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"{{.ModuleTitle}} ID"
//	@Param			body	body		dto.Update{{.ModuleTitle}}Request	true	"Update Request"
//	@Success		200		{object}	response.MyGoResponse{data=dto.{{.ModuleTitle}}Response}
//	@Router			{{.URLPrefix}}/{id} [put]
//
// Update handles PUT {{.URLPrefix}}/:id
func (h *{{.ModuleTitle}}Handler) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	var req dto.Update{{.ModuleTitle}}Request
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, false, "Request tidak valid", nil, nil)
	}

	if errs := validator.Validate(req); errs != nil {
		return response.Response(c, http.StatusUnprocessableEntity, false, "Validasi gagal", nil, errs)
	}

	actor := buildAuthContext(c)
	item, err := h.service.Update(id, &req, getActorID(c), actor)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "{{.ModuleTitle}} tidak ditemukan" {
			status = http.StatusNotFound
		}
		return response.Response(c, status, false, err.Error(), nil, nil)
	}

	return response.Response(c, http.StatusOK, true, "Data berhasil diupdate", item, nil)
}

// ─── Delete ────────────────────────────────────────────────────────────────────
// {{.ModuleTitle}}Handler godoc
//
//	@Summary		Delete {{.ModuleTitle}}
//	@Description	Delete {{.ModuleTitle}} by :id
//	@Tags			{{.MainModule}}/{{.SubModule}}
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"{{.ModuleTitle}} ID"
//	@Success		200		{object}	response.MyGoResponse{}
//	@Router			{{.URLPrefix}}/{id} [delete]
//
// Delete handles DELETE {{.URLPrefix}}/:id
func (h *{{.ModuleTitle}}Handler) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, false, "ID tidak valid", nil, nil)
	}

	actor := buildAuthContext(c)
	if err := h.service.Delete(id, actor); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "{{.ModuleTitle}} tidak ditemukan" {
			status = http.StatusNotFound
		}
		return response.Response(c, status, false, err.Error(), nil, nil)
	}

	return response.Response(c, http.StatusOK, true, "Data berhasil dihapus", nil, nil)
}
`

// Migration-----------------------------------------------------------
var tmplMigration = `package migrations

import (
	"database/sql"
	_ "embed"
	"log"

	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"

	"gorm.io/gorm"
)

//go:embed 001_create_{{.TableName}}_table.sql
var {{.PackageName}}SQL string

// Migrate{{.ModuleTitle}} menjalankan GORM auto-migration
func Migrate{{.ModuleTitle}}(db *gorm.DB) error {
	return db.Migrator().CreateTable(&models.{{.ModuleTitle}}{})
}

// Migrate{{.ModuleTitle}}WithSQL menjalankan migrasi via raw SQL
func Migrate{{.ModuleTitle}}WithSQL(sqlDB *sql.DB) error {
	_, err := sqlDB.Exec({{.PackageName}}SQL)
	if err != nil {
		log.Printf("Error creating {{.TableName}} table: %v", err)
		return err
	}
	log.Println("{{.TableName}} table migrated successfully")
	return nil
}

// Drop{{.ModuleTitle}}Table menghapus tabel (gunakan dengan hati-hati!)
func Drop{{.ModuleTitle}}Table(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.{{.ModuleTitle}}{})
}
`

// Sql-----------------------------------------------------------------
var tmplSQL = `-- Migration: Create {{.TableName}} table
-- Timestamp: {{.Timestamp}}

CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_by  BIGINT,
    updated_by  BIGINT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_{{.TableName}}_deleted_at ON {{.TableName}}(deleted_at);
`

// Factories-----------------------------------------------------------
var tmplFactory = `package factories

import (
	"fmt"
	"math/rand"
	"time"

	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// {{.ModuleTitle}}Factory membuat data {{.ModuleTitle}} untuk testing/seeding
type {{.ModuleTitle}}Factory struct {
	overrides map[string]interface{}
}

func New{{.ModuleTitle}}Factory() *{{.ModuleTitle}}Factory {
	return &{{.ModuleTitle}}Factory{overrides: make(map[string]interface{})}
}

func (f *{{.ModuleTitle}}Factory) With(field string, value interface{}) *{{.ModuleTitle}}Factory {
	f.overrides[field] = value
	return f
}

func (f *{{.ModuleTitle}}Factory) Make() *models.{{.ModuleTitle}} {
	idx := rng.Intn(999999)
	name := fmt.Sprintf("{{.ModuleTitle}} %d", idx)
	desc := fmt.Sprintf("Deskripsi {{.ModuleTitle}} %d", idx)

	if v, ok := f.overrides["name"]; ok {
		name = v.(string)
	}

	return &models.{{.ModuleTitle}}{
		Name:        name,
		Description: &desc,
	}
}

func (f *{{.ModuleTitle}}Factory) MakeMany(count int) []*models.{{.ModuleTitle}} {
	items := make([]*models.{{.ModuleTitle}}, count)
	for i := 0; i < count; i++ {
		items[i] = New{{.ModuleTitle}}Factory().Make()
	}
	return items
}
`

// Seeders-------------------------------------------------------------
var tmplSeeder = `package seeders

import (
	"log"

	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/tests/factories"

	"gorm.io/gorm"
)

// {{.ModuleTitle}}Seeder mengelola seeding data {{.ModuleTitle}}
type {{.ModuleTitle}}Seeder struct {
	db *gorm.DB
}

func New{{.ModuleTitle}}Seeder(db *gorm.DB) *{{.ModuleTitle}}Seeder {
	return &{{.ModuleTitle}}Seeder{db: db}
}

// Run menjalankan seeder
func (s *{{.ModuleTitle}}Seeder) Run() error {
	log.Println("🌱 Seeding {{.TableName}}...")

	items := factories.New{{.ModuleTitle}}Factory().MakeMany(10)
	for _, item := range items {
		if err := s.db.Create(item).Error; err != nil {
			log.Printf("   ⚠️  Gagal membuat {{.ModuleTitle}}: %v", err)
			continue
		}
		log.Printf("   ✅ {{.ModuleTitle}} '%s' dibuat.", item.Name)
	}

	log.Println("✅ {{.TableName}} seeding selesai!")
	return nil
}

// Fresh menghapus semua data lalu seed ulang
func (s *{{.ModuleTitle}}Seeder) Fresh() error {
	log.Println("🗑️  Menghapus semua data {{.TableName}}...")

	if err := s.db.Exec("DELETE FROM {{.TableName}}").Error; err != nil {
		return err
	}
	if err := s.db.Exec("ALTER SEQUENCE {{.TableName}}_id_seq RESTART WITH 1").Error; err != nil {
		log.Printf("Warning: Gagal reset sequence: %v", err)
	}
	return s.Run()
}

// seedDefault menyimpan satu item jika belum ada
func (s *{{.ModuleTitle}}Seeder) seedDefault(name string) error {
	var count int64
	s.db.Model(&models.{{.ModuleTitle}}{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		log.Printf("   ⏭️  '%s' sudah ada, skip.", name)
		return nil
	}

	item := factories.New{{.ModuleTitle}}Factory().With("name", name).Make()
	if err := s.db.Create(item).Error; err != nil {
		return err
	}

	log.Printf("   ✅ '%s' dibuat.", name)
	return nil
}
`

// Helpers-------------------------------------------------------------
var tmplDBHelper = `package helpers

import (
	"log"

	"{{.ProjectModule}}/config"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"

	"gorm.io/gorm"
)

// SetupTestDB membuat koneksi DB untuk keperluan test
func SetupTestDB() *gorm.DB {
	cfg := config.LoadConfig("DEV")
	db, err := cfg.ConnectDB()
	if err != nil {
		log.Fatal("Gagal koneksi DB untuk test:", err)
	}
	return db
}

// MigrateTestDB menjalankan migrasi untuk test DB
func MigrateTestDB(db *gorm.DB) {
	if err := db.AutoMigrate(&models.{{.ModuleTitle}}{}); err != nil {
		log.Fatal("Gagal migrasi test DB:", err)
	}
}

// TruncateTable menghapus semua record dan reset sequence
func TruncateTable(db *gorm.DB, tables ...string) {
	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error; err != nil {
			log.Printf("Warning: Gagal truncate table %s: %v", table, err)
		}
	}
}
`

// Module--------------------------------------------------------------
var tmplModule = `package {{.PackageName}}

import (
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/contracts"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/handlers"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/repositories"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/services"
	"{{.ProjectModule}}/internal/shared/utils"

	authContracts "{{.ProjectModule}}/internal/modules/auth/contracts" //auth
	rbacContracts "{{.ProjectModule}}/internal/modules/rbac/contracts" //rbac

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Module mewakili {{.MainModule}}/{{.SubModule}} module
type Module struct {
	db         *gorm.DB
	handler    *handlers.{{.ModuleTitle}}Handler
	jwtManager *utils.JWTManager
	repo       contracts.Repository
	rbacRepo   rbacContracts.RBACRepository //RBAC
}

// NewModule membuat instance module baru dan wire semua layer
func NewModule(
	db *gorm.DB, 
	jwtManager *utils.JWTManager,
	rbacRepo rbacContracts.RBACRepository, //RBAC
	authRepo authContracts.AuthRepository, //AUTH
) *Module {
	repo := repositories.New{{.ModuleTitle}}Repository(db)
	svc := services.New{{.ModuleTitle}}Service(
		repo,
		rbacRepo,
		authRepo,
	)
	handler := handlers.New{{.ModuleTitle}}Handler(svc)

	return &Module{
		db:         db,
		handler:    handler,
		jwtManager: jwtManager,
		repo:       repo,
		rbacRepo:   rbacRepo, //RBAC
	}
}

// InitRoutes mendaftarkan routes ke echo instance
func (m *Module) InitRoutes(e *echo.Echo) {
	RegisterRoutes(e, m.handler, m.jwtManager, m.db)
}
`

// Route---------------------------------------------------------------
var tmplRoutes = `package {{.PackageName}}

import (
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/handlers"
	authMiddlewares "{{.ProjectModule}}/internal/modules/auth/middlewares"
	"{{.ProjectModule}}/internal/shared/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
	
)

// RegisterRoutes mendaftarkan semua routes untuk module {{.MainModule}}/{{.SubModule}}
func RegisterRoutes(e *echo.Echo, h *handlers.{{.ModuleTitle}}Handler, jwtManager *utils.JWTManager, db *gorm.DB) {
	jwt := authMiddlewares.JWTMiddleware(jwtManager, db)
	g := e.Group("{{.URLPrefix}}", jwt)
	g.GET("", h.List)
	g.GET("/:id", h.GetByID)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
`

// Register------------------------------------------------------------
var tmplRegister = `package {{.PackageName}}

import (
	"database/sql"

	"{{.ProjectModule}}/config"
	"{{.ProjectModule}}/internal/apps"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/migrations"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
	"{{.ProjectModule}}/internal/shared/utils"

	authContracts "{{.ProjectModule}}/internal/modules/auth/contracts"
	authRepositories "{{.ProjectModule}}/internal/modules/auth/repositories"
	rbacContracts "{{.ProjectModule}}/internal/modules/rbac/contracts"
	rbacRepositories "{{.ProjectModule}}/internal/modules/rbac/repositories"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type registryModule struct {
	db  *gorm.DB
	cfg *config.Config
	rbacRepo rbacContracts.RBACRepository //RBAC
	authRepo authContracts.AuthRepository //AUTH
}

// init dipanggil otomatis saat package di-import (blank import)
func init() {
	apps.Register(&registryModule{})
}

func (r *registryModule) SetDB(db *gorm.DB)            { 
	r.db = db 
	r.rbacRepo = rbacRepositories.NewRBACRepository(db) //RBAC
	r.authRepo = authRepositories.NewAuthRepository(db) //AUTH
}
func (r *registryModule) SetConfig(cfg *config.Config) { 
	r.cfg = cfg
}

func (r *registryModule) InitRoutes(e *echo.Echo) {
	jwtManager := utils.NewJWTManager(
		r.cfg.JWTSecret,
		r.cfg.JWTIssuer,
		r.cfg.JWTAccessTokenExpMinutes,
		r.cfg.JWTRefreshTokenExpDays,
	)
	NewModule(r.db, jwtManager, r.rbacRepo, r.authRepo).InitRoutes(e)
}

func (r *registryModule) Models() []interface{} {
	return []interface{}{
		&models.{{.ModuleTitle}}{},
	}
}

func (r *registryModule) SeedData(db *gorm.DB) error {
	return nil
}

func (r *registryModule) MigrateSQL(sqlDB *sql.DB) error {
	return migrations.Migrate{{.ModuleTitle}}WithSQL(sqlDB)
}
`

// Tests---------------------------------------------------------------
// -----Mocks----------------------------------------------------------
// -----------Module Service Mock--------------------------------------

var tmplModuleServiceMock = `package mocks

import (
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/dto"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
	"github.com/stretchr/testify/mock"
)

// {{.ModuleTitle}}RepositoryMock is a mock implementation of contracts.Repository
type {{.ModuleTitle}}RepositoryMock struct {
	mock.Mock
}

func (m *{{.ModuleTitle}}RepositoryMock) Create(item *models.{{.ModuleTitle}}) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *{{.ModuleTitle}}RepositoryMock) GetByID(id int64) (*models.{{.ModuleTitle}}, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.{{.ModuleTitle}}), args.Error(1)
}

func (m *{{.ModuleTitle}}RepositoryMock) List(page, pageSize int, filter *dto.Filter{{.ModuleTitle}}Request) ([]models.{{.ModuleTitle}}, int64, error) {
	args := m.Called(page, pageSize, filter)
	return args.Get(0).([]models.{{.ModuleTitle}}), args.Get(1).(int64), args.Error(2)
}

func (m *{{.ModuleTitle}}RepositoryMock) Update(item *models.{{.ModuleTitle}}) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *{{.ModuleTitle}}RepositoryMock) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}
`

var tmplRBACMock = `package mocks

import (
	rbacModels "{{.ProjectModule}}/internal/modules/rbac/models"
	"github.com/stretchr/testify/mock"
)

// RBACRepositoryMock is a mock implementation of rbacContracts.RBACRepository
type RBACRepositoryMock struct {
	mock.Mock
}

func (m *RBACRepositoryMock) IsSuperadmin(userID int64) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *RBACRepositoryMock) CreatePermission(p *rbacModels.Permission) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *RBACRepositoryMock) GetPermissionByID(id int64) (*rbacModels.Permission, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rbacModels.Permission), args.Error(1)
}

func (m *RBACRepositoryMock) GetPermissionByName(name string) (*rbacModels.Permission, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rbacModels.Permission), args.Error(1)
}

func (m *RBACRepositoryMock) ListPermissions(page, pageSize int) ([]rbacModels.Permission, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]rbacModels.Permission), args.Get(1).(int64), args.Error(2)
}

func (m *RBACRepositoryMock) UpdatePermission(p *rbacModels.Permission) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *RBACRepositoryMock) DeletePermission(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *RBACRepositoryMock) CreateRole(r *rbacModels.Role) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *RBACRepositoryMock) GetRoleByID(id int64) (*rbacModels.Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rbacModels.Role), args.Error(1)
}

func (m *RBACRepositoryMock) GetRoleByName(name string) (*rbacModels.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rbacModels.Role), args.Error(1)
}

func (m *RBACRepositoryMock) ListRoles(page, pageSize int) ([]rbacModels.Role, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]rbacModels.Role), args.Get(1).(int64), args.Error(2)
}

func (m *RBACRepositoryMock) UpdateRole(r *rbacModels.Role) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *RBACRepositoryMock) DeleteRole(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *RBACRepositoryMock) GetUsersRoles(userIDs []int64) (map[int64][]rbacModels.Role, error) {
	args := m.Called(userIDs)
	return args.Get(0).(map[int64][]rbacModels.Role), args.Error(1)
}

func (m *RBACRepositoryMock) AssignPermissionsToRole(roleID int64, permissionIDs []int64) error {
	args := m.Called(roleID, permissionIDs)
	return args.Error(0)
}

func (m *RBACRepositoryMock) RevokePermissionsFromRole(roleID int64, permissionIDs []int64) error {
	args := m.Called(roleID, permissionIDs)
	return args.Error(0)
}

func (m *RBACRepositoryMock) GetRolePermissions(roleID int64) ([]rbacModels.Permission, error) {
	args := m.Called(roleID)
	return args.Get(0).([]rbacModels.Permission), args.Error(1)
}

func (m *RBACRepositoryMock) SyncRolePermissions(roleID int64, permissionIDs []int64) error {
	args := m.Called(roleID, permissionIDs)
	return args.Error(0)
}

func (m *RBACRepositoryMock) AssignRolesToUser(userID int64, roleIDs []int64, assignedBy *int64) error {
	args := m.Called(userID, roleIDs, assignedBy)
	return args.Error(0)
}

func (m *RBACRepositoryMock) RevokeRolesFromUser(userID int64, roleIDs []int64) error {
	args := m.Called(userID, roleIDs)
	return args.Error(0)
}

func (m *RBACRepositoryMock) GetUserRoles(userID int64) ([]rbacModels.Role, error) {
	args := m.Called(userID)
	return args.Get(0).([]rbacModels.Role), args.Error(1)
}

func (m *RBACRepositoryMock) SyncUserRoles(userID int64, roleIDs []int64, assignedBy *int64) error {
	args := m.Called(userID, roleIDs, assignedBy)
	return args.Error(0)
}

func (m *RBACRepositoryMock) AssignDirectPermission(userID, permissionID int64, isGranted bool, assignedBy *int64) error {
	args := m.Called(userID, permissionID, isGranted, assignedBy)
	return args.Error(0)
}

func (m *RBACRepositoryMock) RevokeDirectPermission(userID, permissionID int64) error {
	args := m.Called(userID, permissionID)
	return args.Error(0)
}

func (m *RBACRepositoryMock) GetUserDirectPermissions(userID int64) ([]rbacModels.UserPermission, error) {
	args := m.Called(userID)
	return args.Get(0).([]rbacModels.UserPermission), args.Error(1)
}

func (m *RBACRepositoryMock) GetUserAllPermissions(userID int64) ([]string, error) {
	args := m.Called(userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *RBACRepositoryMock) HasPermission(userID int64, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}
`

var tmplAuthMock = `package mocks

import (
	authModels "{{.ProjectModule}}/internal/modules/auth/models"
	"github.com/stretchr/testify/mock"
)

// AuthRepositoryMock is a mock implementation of authContracts.AuthRepository
type AuthRepositoryMock struct {
	mock.Mock
}

func (m *AuthRepositoryMock) SaveToken(token *authModels.AuthToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *AuthRepositoryMock) GetTokenByJTI(jti string) (*authModels.AuthToken, error) {
	args := m.Called(jti)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authModels.AuthToken), args.Error(1)
}

func (m *AuthRepositoryMock) BlacklistToken(jti string) error {
	args := m.Called(jti)
	return args.Error(0)
}

func (m *AuthRepositoryMock) BlacklistAllUserTokens(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *AuthRepositoryMock) CountActiveTokens(userID int64) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *AuthRepositoryMock) SaveLoginHistory(history *authModels.LoginHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *AuthRepositoryMock) GetUserLoginHistories(userID int64, limit int) ([]authModels.LoginHistory, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]authModels.LoginHistory), args.Error(1)
}

func (m *AuthRepositoryMock) SavePasswordHistory(history *authModels.PasswordHistory) error {
	args := m.Called(history)
	return args.Error(0)
}

func (m *AuthRepositoryMock) GetPasswordHistories(userID int64, limit int) ([]authModels.PasswordHistory, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]authModels.PasswordHistory), args.Error(1)
}
`

var tmplModuleServiceTest = `package tests

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/dto"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/models"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/services"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/tests/factories"
	"{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/tests/mocks"

	{{.ModuleName}}Contracts "{{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}/contracts"
	rbacModels "{{.ProjectModule}}/internal/modules/rbac/models"
	appErrors "{{.ProjectModule}}/internal/shared/errors"
)

func TestMain(m *testing.M) {
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")
	fmt.Println("\033[35m  {{.ModuleTitle}} Service Test Suite\033[0m")
	fmt.Println("\033[34m" + strings.Repeat("─", 55) + "\033[0m")

	code := m.Run()

	if code == 0 {
		fmt.Println("\n\033[32m✓  PASS\033[0m  {{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}")
	} else {
		fmt.Println("\n\033[31m✗  FAIL\033[0m  {{.ProjectModule}}/internal/modules/{{.MainModule}}/{{.SubModule}}")
	}

	os.Exit(code)
}

type {{.ModuleTitle}}ServiceTestSuite struct {
	suite.Suite
	repo     *mocks.{{.ModuleTitle}}RepositoryMock
	rbacRepo *mocks.RBACRepositoryMock
	authRepo *mocks.AuthRepositoryMock
	svc      {{.ModuleName}}Contracts.Service
}

func (s *{{.ModuleTitle}}ServiceTestSuite) SetupTest() {
	s.repo     = new(mocks.{{.ModuleTitle}}RepositoryMock)
	s.rbacRepo = new(mocks.RBACRepositoryMock)
	s.authRepo = new(mocks.AuthRepositoryMock)
	s.svc = services.New{{.ModuleTitle}}Service(s.repo, s.rbacRepo, s.authRepo)
}

func Test{{.ModuleTitle}}Service(t *testing.T) {
	suite.Run(t, new({{.ModuleTitle}}ServiceTestSuite))
}

func superadminActor() {{.ModuleName}}Contracts.AuthContext {
	return {{.ModuleName}}Contracts.AuthContext{UserID: 1, IsSuperadmin: true}
}

func regularActor() {{.ModuleName}}Contracts.AuthContext {
	return {{.ModuleName}}Contracts.AuthContext{UserID: 2, IsSuperadmin: false}
}

func (s *{{.ModuleTitle}}ServiceTestSuite) mockHasPermission(perm string, result bool) {
	s.rbacRepo.On("HasPermission", regularActor().UserID, perm, mock.Anything).Return(result, nil).Maybe()
}

func (s *{{.ModuleTitle}}ServiceTestSuite) mockNoPermissions() {
	s.rbacRepo.On("HasPermission", regularActor().UserID, mock.Anything, mock.Anything).Return(false, nil)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Create_Superadmin_Success() {
	req := &dto.Create{{.ModuleTitle}}Request{
		Name: "Test {{.ModuleTitle}}",
	}
	actor := superadminActor()

	s.repo.On("Create", mock.AnythingOfType("*models.{{.ModuleTitle}}")).Return(nil)

	result, err := s.svc.Create(req, &actor.UserID, actor)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(req.Name, result.Name)
	s.repo.AssertExpectations(s.T())
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Create_WithPermission_Success() {
	req := &dto.Create{{.ModuleTitle}}Request{
		Name: "Test {{.ModuleTitle}}",
	}
	actor := regularActor()

	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermAnyCreate).Return(true, nil)
	s.repo.On("Create", mock.AnythingOfType("*models.{{.ModuleTitle}}")).Return(nil)

	result, err := s.svc.Create(req, &actor.UserID, actor)

	s.NoError(err)
	s.NotNil(result)
	s.repo.AssertExpectations(s.T())
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Create_WithManagePermission_Success() {
	req := &dto.Create{{.ModuleTitle}}Request{Name: "Test"}
	actor := regularActor()

	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermAnyCreate).Return(false, nil)
	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermAnyManage).Return(true, nil)
	s.repo.On("Create", mock.AnythingOfType("*models.{{.ModuleTitle}}")).Return(nil)

	result, err := s.svc.Create(req, &actor.UserID, actor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Create_Forbidden() {
	req := &dto.Create{{.ModuleTitle}}Request{Name: "Test"}
	actor := regularActor()
	s.mockNoPermissions()

	result, err := s.svc.Create(req, &actor.UserID, actor)

	s.Nil(result)
	s.Error(err)
	var appErr *appErrors.AppError
	s.ErrorAs(err, &appErr)
	s.Equal(http.StatusForbidden, appErr.Code)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Create_RepoError() {
	req := &dto.Create{{.ModuleTitle}}Request{Name: "Test"}
	actor := superadminActor()

	s.repo.On("Create", mock.AnythingOfType("*models.{{.ModuleTitle}}")).Return(fmt.Errorf("db error"))

	result, err := s.svc.Create(req, &actor.UserID, actor)

	s.Nil(result)
	s.Error(err)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_GetByID_Superadmin_Success() {
	actor := superadminActor()
	item := factories.New{{.ModuleTitle}}Factory().Make()
	item.ID = 1

	s.repo.On("GetByID", int64(1)).Return(item, nil)

	result, err := s.svc.GetByID(1, actor)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(item.ID, result.ID)
	s.Equal(item.Name, result.Name)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_GetByID_WithPermission_Success() {
	actor := regularActor()
	item := factories.New{{.ModuleTitle}}Factory().Make()
	item.ID = 1

	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermAnyRead).Return(true, nil)
	s.repo.On("GetByID", int64(1)).Return(item, nil)

	result, err := s.svc.GetByID(1, actor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_GetByID_Forbidden() {
	actor := regularActor()
	s.mockNoPermissions()

	result, err := s.svc.GetByID(1, actor)

	s.Nil(result)
	s.Error(err)
	var appErr *appErrors.AppError
	s.ErrorAs(err, &appErr)
	s.Equal(http.StatusForbidden, appErr.Code)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_GetByID_NotFound() {
	actor := superadminActor()

	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	result, err := s.svc.GetByID(999, actor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_GetByID_RepoError() {
	actor := superadminActor()

	s.repo.On("GetByID", int64(1)).Return(nil, fmt.Errorf("db error"))

	result, err := s.svc.GetByID(1, actor)

	s.Nil(result)
	s.Error(err)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_List_Superadmin_Success() {
	actor := superadminActor()
	filter := &dto.Filter{{.ModuleTitle}}Request{}
	items := []models.{{.ModuleTitle}}{
		*factories.New{{.ModuleTitle}}Factory().Make(),
		*factories.New{{.ModuleTitle}}Factory().Make(),
	}

	s.repo.On("List", 1, 10, filter).Return(items, int64(2), nil)

	result, total, err := s.svc.List(1, 10, filter, actor)

	s.NoError(err)
	s.Equal(int64(2), total)
	s.Len(result, 2)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_List_WithPermission_Success() {
	actor := regularActor()
	filter := &dto.Filter{{.ModuleTitle}}Request{}
	items := []models.{{.ModuleTitle}}{*factories.New{{.ModuleTitle}}Factory().Make()}

	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermAnyRead).Return(true, nil)
	s.repo.On("List", 1, 10, filter).Return(items, int64(1), nil)

	result, total, err := s.svc.List(1, 10, filter, actor)

	s.NoError(err)
	s.Equal(int64(1), total)
	s.Len(result, 1)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_List_Forbidden() {
	actor := regularActor()
	filter := &dto.Filter{{.ModuleTitle}}Request{}
	s.mockNoPermissions()

	result, total, err := s.svc.List(1, 10, filter, actor)

	s.Nil(result)
	s.Equal(int64(0), total)
	s.Error(err)
	var appErr *appErrors.AppError
	s.ErrorAs(err, &appErr)
	s.Equal(http.StatusForbidden, appErr.Code)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_List_DefaultPagination() {
	actor := superadminActor()
	filter := &dto.Filter{{.ModuleTitle}}Request{}

	s.repo.On("List", 1, 10, filter).Return([]models.{{.ModuleTitle}}{}, int64(0), nil)

	result, total, err := s.svc.List(0, 0, filter, actor)

	s.NoError(err)
	s.Equal(int64(0), total)
	s.Empty(result)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_List_PageSizeCapped() {
	actor := superadminActor()
	filter := &dto.Filter{{.ModuleTitle}}Request{}

	s.repo.On("List", 1, 10, filter).Return([]models.{{.ModuleTitle}}{}, int64(0), nil)

	_, _, err := s.svc.List(1, 999, filter, actor)

	s.NoError(err)
	s.repo.AssertCalled(s.T(), "List", 1, 10, filter)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_List_WithNameFilter() {
	actor := superadminActor()
	filter := &dto.Filter{{.ModuleTitle}}Request{Name: "test"}
	items := []models.{{.ModuleTitle}}{*factories.New{{.ModuleTitle}}Factory().Make()}

	s.repo.On("List", 1, 10, filter).Return(items, int64(1), nil)

	result, total, err := s.svc.List(1, 10, filter, actor)

	s.NoError(err)
	s.Equal(int64(1), total)
	s.Len(result, 1)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Update_Superadmin_Success() {
	actor := superadminActor()
	existing := factories.New{{.ModuleTitle}}Factory().Make()
	existing.ID = 1
	newName := "Updated Name"
	req := &dto.Update{{.ModuleTitle}}Request{Name: &newName}

	s.repo.On("GetByID", int64(1)).Return(existing, nil)
	s.repo.On("Update", mock.AnythingOfType("*models.{{.ModuleTitle}}")).Return(nil)

	result, err := s.svc.Update(1, req, &actor.UserID, actor)

	s.NoError(err)
	s.NotNil(result)
	s.Equal(newName, result.Name)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Update_WithPermission_Success() {
	actor := regularActor()
	existing := factories.New{{.ModuleTitle}}Factory().Make()
	existing.ID = 1
	newName := "Updated"
	req := &dto.Update{{.ModuleTitle}}Request{Name: &newName}

	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermAnyUpdate).Return(true, nil)
	s.repo.On("GetByID", int64(1)).Return(existing, nil)
	s.repo.On("Update", mock.AnythingOfType("*models.{{.ModuleTitle}}")).Return(nil)

	result, err := s.svc.Update(1, req, &actor.UserID, actor)

	s.NoError(err)
	s.NotNil(result)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Update_Forbidden() {
	actor := regularActor()
	req := &dto.Update{{.ModuleTitle}}Request{}
	s.mockNoPermissions()

	result, err := s.svc.Update(1, req, &actor.UserID, actor)

	s.Nil(result)
	s.Error(err)
	var appErr *appErrors.AppError
	s.ErrorAs(err, &appErr)
	s.Equal(http.StatusForbidden, appErr.Code)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Update_NotFound() {
	actor := superadminActor()
	req := &dto.Update{{.ModuleTitle}}Request{}

	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	result, err := s.svc.Update(999, req, &actor.UserID, actor)

	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Update_PartialFields() {
	actor := superadminActor()
	existing := factories.New{{.ModuleTitle}}Factory().Make()
	existing.ID = 1
	originalName := existing.Name
	newDesc := "New description"

	req := &dto.Update{{.ModuleTitle}}Request{Description: &newDesc}

	s.repo.On("GetByID", int64(1)).Return(existing, nil)
	s.repo.On("Update", mock.MatchedBy(func(m *models.{{.ModuleTitle}}) bool {
		return m.Name == originalName && *m.Description == newDesc
	})).Return(nil)

	result, err := s.svc.Update(1, req, &actor.UserID, actor)

	s.NoError(err)
	s.Equal(originalName, result.Name)
	s.Equal(newDesc, *result.Description)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Update_RepoError() {
	actor := superadminActor()
	existing := factories.New{{.ModuleTitle}}Factory().Make()
	existing.ID = 1
	req := &dto.Update{{.ModuleTitle}}Request{}

	s.repo.On("GetByID", int64(1)).Return(existing, nil)
	s.repo.On("Update", mock.AnythingOfType("*models.{{.ModuleTitle}}")).Return(fmt.Errorf("db error"))

	result, err := s.svc.Update(1, req, &actor.UserID, actor)

	s.Nil(result)
	s.Error(err)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Delete_Superadmin_Success() {
	actor := superadminActor()
	existing := factories.New{{.ModuleTitle}}Factory().Make()
	existing.ID = 1

	s.repo.On("GetByID", int64(1)).Return(existing, nil)
	s.repo.On("Delete", int64(1)).Return(nil)

	err := s.svc.Delete(1, actor)

	s.NoError(err)
	s.repo.AssertExpectations(s.T())
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Delete_WithPermission_Success() {
	actor := regularActor()
	existing := factories.New{{.ModuleTitle}}Factory().Make()
	existing.ID = 1

	s.rbacRepo.On("HasPermission", actor.UserID, rbacModels.PermAnyDelete).Return(true, nil)
	s.repo.On("GetByID", int64(1)).Return(existing, nil)
	s.repo.On("Delete", int64(1)).Return(nil)

	err := s.svc.Delete(1, actor)

	s.NoError(err)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Delete_Forbidden() {
	actor := regularActor()
	s.mockNoPermissions()

	err := s.svc.Delete(1, actor)

	s.Error(err)
	var appErr *appErrors.AppError
	s.ErrorAs(err, &appErr)
	s.Equal(http.StatusForbidden, appErr.Code)
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Delete_NotFound() {
	actor := superadminActor()

	s.repo.On("GetByID", int64(999)).Return(nil, nil)

	err := s.svc.Delete(999, actor)

	s.Error(err)
	s.Contains(err.Error(), "tidak ditemukan")
}

func (s *{{.ModuleTitle}}ServiceTestSuite) Test_Delete_RepoError() {
	actor := superadminActor()
	existing := factories.New{{.ModuleTitle}}Factory().Make()
	existing.ID = 1

	s.repo.On("GetByID", int64(1)).Return(existing, nil)
	s.repo.On("Delete", int64(1)).Return(fmt.Errorf("db error"))

	err := s.svc.Delete(1, actor)

	s.Error(err)
}
`
