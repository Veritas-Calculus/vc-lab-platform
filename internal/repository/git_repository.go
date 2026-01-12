// Package repository provides data access implementations.
package repository

import (
	"context"
	"errors"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// GitRepoRepository defines the interface for git repository data access.
type GitRepoRepository interface {
	Create(ctx context.Context, repo *model.GitRepository) error
	GetByID(ctx context.Context, id string) (*model.GitRepository, error)
	GetByType(ctx context.Context, repoType model.GitRepoType) ([]model.GitRepository, error)
	GetDefaultByType(ctx context.Context, repoType model.GitRepoType) (*model.GitRepository, error)
	List(ctx context.Context, page, pageSize int) ([]model.GitRepository, int64, error)
	Update(ctx context.Context, repo *model.GitRepository) error
	Delete(ctx context.Context, id string) error
}

type gitRepoRepository struct {
	db *gorm.DB
}

// NewGitRepoRepository creates a new git repository repository.
func NewGitRepoRepository(db *gorm.DB) GitRepoRepository {
	return &gitRepoRepository{db: db}
}

func (r *gitRepoRepository) Create(ctx context.Context, repo *model.GitRepository) error {
	return r.db.WithContext(ctx).Create(repo).Error
}

func (r *gitRepoRepository) GetByID(ctx context.Context, id string) (*model.GitRepository, error) {
	var repo model.GitRepository
	if err := r.db.WithContext(ctx).First(&repo, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &repo, nil
}

func (r *gitRepoRepository) GetByType(ctx context.Context, repoType model.GitRepoType) ([]model.GitRepository, error) {
	var repos []model.GitRepository
	if err := r.db.WithContext(ctx).
		Where("type = ? AND status = ?", repoType, 1).
		Order("created_at DESC").
		Find(&repos).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

func (r *gitRepoRepository) GetDefaultByType(ctx context.Context, repoType model.GitRepoType) (*model.GitRepository, error) {
	var repo model.GitRepository
	if err := r.db.WithContext(ctx).
		Where("type = ? AND is_default = ? AND status = ?", repoType, true, 1).
		First(&repo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &repo, nil
}

func (r *gitRepoRepository) List(ctx context.Context, page, pageSize int) ([]model.GitRepository, int64, error) {
	var repos []model.GitRepository
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.GitRepository{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&repos).Error; err != nil {
		return nil, 0, err
	}

	return repos, total, nil
}

func (r *gitRepoRepository) Update(ctx context.Context, repo *model.GitRepository) error {
	return r.db.WithContext(ctx).Save(repo).Error
}

func (r *gitRepoRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.GitRepository{}, "id = ?", id).Error
}

// NodeConfigRepository defines the interface for node config data access.
type NodeConfigRepository interface {
	Create(ctx context.Context, config *model.NodeConfig) error
	GetByID(ctx context.Context, id string) (*model.NodeConfig, error)
	GetByResourceRequestID(ctx context.Context, requestID string) (*model.NodeConfig, error)
	ListByStorageRepo(ctx context.Context, repoID string, page, pageSize int) ([]model.NodeConfig, int64, error)
	ListByStatus(ctx context.Context, status model.NodeConfigStatus, page, pageSize int) ([]model.NodeConfig, int64, error)
	Update(ctx context.Context, config *model.NodeConfig) error
	Delete(ctx context.Context, id string) error
}

type nodeConfigRepository struct {
	db *gorm.DB
}

// NewNodeConfigRepository creates a new node config repository.
func NewNodeConfigRepository(db *gorm.DB) NodeConfigRepository {
	return &nodeConfigRepository{db: db}
}

func (r *nodeConfigRepository) Create(ctx context.Context, config *model.NodeConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *nodeConfigRepository) GetByID(ctx context.Context, id string) (*model.NodeConfig, error) {
	var config model.NodeConfig
	if err := r.db.WithContext(ctx).
		Preload("ResourceRequest").
		Preload("StorageRepo").
		Preload("ModuleRepo").
		First(&config, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &config, nil
}

func (r *nodeConfigRepository) GetByResourceRequestID(ctx context.Context, requestID string) (*model.NodeConfig, error) {
	var config model.NodeConfig
	if err := r.db.WithContext(ctx).
		Preload("ResourceRequest").
		Preload("StorageRepo").
		Preload("ModuleRepo").
		Where("resource_request_id = ?", requestID).
		First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &config, nil
}

func (r *nodeConfigRepository) ListByStorageRepo(ctx context.Context, repoID string, page, pageSize int) ([]model.NodeConfig, int64, error) {
	var configs []model.NodeConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.NodeConfig{}).Where("storage_repo_id = ?", repoID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Preload("ResourceRequest").
		Preload("StorageRepo").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

func (r *nodeConfigRepository) ListByStatus(ctx context.Context, status model.NodeConfigStatus, page, pageSize int) ([]model.NodeConfig, int64, error) {
	var configs []model.NodeConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.NodeConfig{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Preload("ResourceRequest").
		Preload("StorageRepo").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

func (r *nodeConfigRepository) Update(ctx context.Context, config *model.NodeConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *nodeConfigRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.NodeConfig{}, "id = ?", id).Error
}
