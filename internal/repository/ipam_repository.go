// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"gorm.io/gorm"
)

// IPPoolRepository defines the interface for IP pool operations.
type IPPoolRepository interface {
	Create(ctx context.Context, pool *model.IPPool) error
	GetByID(ctx context.Context, id string) (*model.IPPool, error)
	List(ctx context.Context, zoneID string, offset, limit int) ([]*model.IPPool, int64, error)
	Update(ctx context.Context, pool *model.IPPool) error
	Delete(ctx context.Context, id string) error
}

// IPAllocationRepository defines the interface for IP allocation operations.
type IPAllocationRepository interface {
	Create(ctx context.Context, allocation *model.IPAllocation) error
	GetByID(ctx context.Context, id string) (*model.IPAllocation, error)
	GetByIPAddress(ctx context.Context, poolID, ipAddress string) (*model.IPAllocation, error)
	ListByPool(ctx context.Context, poolID string, offset, limit int) ([]*model.IPAllocation, int64, error)
	ListByResource(ctx context.Context, resourceID string) ([]*model.IPAllocation, error)
	Update(ctx context.Context, allocation *model.IPAllocation) error
	Delete(ctx context.Context, id string) error
	AllocateNextAvailable(ctx context.Context, poolID, hostname, resourceID string) (*model.IPAllocation, error)
	Release(ctx context.Context, id string) error
	GetAvailableCount(ctx context.Context, poolID string) (int64, error)
}

type ipPoolRepository struct {
	db *gorm.DB
}

type ipAllocationRepository struct {
	db *gorm.DB
}

// NewIPPoolRepository creates a new IP pool repository.
func NewIPPoolRepository(db *gorm.DB) IPPoolRepository {
	return &ipPoolRepository{db: db}
}

// NewIPAllocationRepository creates a new IP allocation repository.
func NewIPAllocationRepository(db *gorm.DB) IPAllocationRepository {
	return &ipAllocationRepository{db: db}
}

// Create creates a new IP pool.
func (r *ipPoolRepository) Create(ctx context.Context, pool *model.IPPool) error {
	return r.db.WithContext(ctx).Create(pool).Error
}

// GetByID retrieves an IP pool by ID.
func (r *ipPoolRepository) GetByID(ctx context.Context, id string) (*model.IPPool, error) {
	var pool model.IPPool
	if err := r.db.WithContext(ctx).Preload("Zone").First(&pool, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &pool, nil
}

// List retrieves IP pools with optional zone filtering.
func (r *ipPoolRepository) List(ctx context.Context, zoneID string, offset, limit int) ([]*model.IPPool, int64, error) {
	var pools []*model.IPPool
	var total int64

	query := r.db.WithContext(ctx).Model(&model.IPPool{})
	if zoneID != "" {
		query = query.Where("zone_id = ?", zoneID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Zone").Offset(offset).Limit(limit).Order("created_at DESC").Find(&pools).Error; err != nil {
		return nil, 0, err
	}

	return pools, total, nil
}

// Update updates an existing IP pool.
func (r *ipPoolRepository) Update(ctx context.Context, pool *model.IPPool) error {
	return r.db.WithContext(ctx).Save(pool).Error
}

// Delete deletes an IP pool by ID.
func (r *ipPoolRepository) Delete(ctx context.Context, id string) error {
	// Check if there are any allocations
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.IPAllocation{}).Where("ip_pool_id = ? AND status != ?", id, "available").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete IP pool with active allocations")
	}

	result := r.db.WithContext(ctx).Delete(&model.IPPool{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Create creates a new IP allocation.
func (r *ipAllocationRepository) Create(ctx context.Context, allocation *model.IPAllocation) error {
	return r.db.WithContext(ctx).Create(allocation).Error
}

// GetByID retrieves an IP allocation by ID.
func (r *ipAllocationRepository) GetByID(ctx context.Context, id string) (*model.IPAllocation, error) {
	var allocation model.IPAllocation
	if err := r.db.WithContext(ctx).Preload("IPPool").First(&allocation, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &allocation, nil
}

// GetByIPAddress retrieves an IP allocation by pool ID and IP address.
func (r *ipAllocationRepository) GetByIPAddress(ctx context.Context, poolID, ipAddress string) (*model.IPAllocation, error) {
	var allocation model.IPAllocation
	if err := r.db.WithContext(ctx).Preload("IPPool").First(&allocation, "ip_pool_id = ? AND ip_address = ?", poolID, ipAddress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &allocation, nil
}

// ListByPool retrieves IP allocations for a specific pool.
func (r *ipAllocationRepository) ListByPool(ctx context.Context, poolID string, offset, limit int) ([]*model.IPAllocation, int64, error) {
	var allocations []*model.IPAllocation
	var total int64

	query := r.db.WithContext(ctx).Model(&model.IPAllocation{}).Where("ip_pool_id = ?", poolID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("IPPool").Offset(offset).Limit(limit).Order("ip_address ASC").Find(&allocations).Error; err != nil {
		return nil, 0, err
	}

	return allocations, total, nil
}

// ListByResource retrieves IP allocations for a specific resource.
func (r *ipAllocationRepository) ListByResource(ctx context.Context, resourceID string) ([]*model.IPAllocation, error) {
	var allocations []*model.IPAllocation
	if err := r.db.WithContext(ctx).Preload("IPPool").Where("resource_id = ?", resourceID).Find(&allocations).Error; err != nil {
		return nil, err
	}
	return allocations, nil
}

// Update updates an existing IP allocation.
func (r *ipAllocationRepository) Update(ctx context.Context, allocation *model.IPAllocation) error {
	return r.db.WithContext(ctx).Save(allocation).Error
}

// Delete deletes an IP allocation by ID.
func (r *ipAllocationRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.IPAllocation{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// AllocateNextAvailable allocates the next available IP address from a pool.
//
//nolint:gocognit // complexity is inherent to transactional IP allocation logic
func (r *ipAllocationRepository) AllocateNextAvailable(ctx context.Context, poolID, hostname, resourceID string) (*model.IPAllocation, error) {
	var allocation *model.IPAllocation

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get the pool
		var pool model.IPPool
		if err := tx.First(&pool, "id = ?", poolID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		// Get all allocated IPs in this pool
		var allocatedIPs []string
		if err := tx.Model(&model.IPAllocation{}).
			Where("ip_pool_id = ? AND status != ?", poolID, "available").
			Pluck("ip_address", &allocatedIPs).Error; err != nil {
			return err
		}

		// Convert to map for quick lookup
		allocatedMap := make(map[string]bool)
		for _, ip := range allocatedIPs {
			allocatedMap[ip] = true
		}

		// Find next available IP
		startIP := net.ParseIP(pool.StartIP)
		endIP := net.ParseIP(pool.EndIP)
		if startIP == nil || endIP == nil {
			return errors.New("invalid IP range in pool")
		}

		var nextIP net.IP
		for ip := dupIP(startIP); !ip.Equal(endIP); incrementIP(ip) {
			ipStr := ip.String()
			if !allocatedMap[ipStr] {
				nextIP = make(net.IP, len(ip))
				copy(nextIP, ip)
				break
			}
		}

		if nextIP == nil {
			// Check the end IP too
			if !allocatedMap[endIP.String()] {
				nextIP = endIP
			} else {
				return errors.New("no available IP addresses in pool")
			}
		}

		// Create the allocation
		now := time.Now()
		var resID *string
		if resourceID != "" {
			resID = &resourceID
		}
		allocation = &model.IPAllocation{
			IPPoolID:    poolID,
			IPAddress:   nextIP.String(),
			Hostname:    hostname,
			ResourceID:  resID,
			Status:      "allocated",
			AllocatedAt: &now,
		}

		return tx.Create(allocation).Error
	})

	if err != nil {
		return nil, err
	}

	return allocation, nil
}

// Release releases an IP allocation back to the pool.
func (r *ipAllocationRepository) Release(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&model.IPAllocation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "available",
			"hostname":     "",
			"resource_id":  "",
			"allocated_at": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetAvailableCount returns the count of available IPs in a pool.
func (r *ipAllocationRepository) GetAvailableCount(ctx context.Context, poolID string) (int64, error) {
	// Get the pool
	var pool model.IPPool
	if err := r.db.WithContext(ctx).First(&pool, "id = ?", poolID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	// Calculate total IPs in range
	startIP := net.ParseIP(pool.StartIP)
	endIP := net.ParseIP(pool.EndIP)
	if startIP == nil || endIP == nil {
		return 0, errors.New("invalid IP range in pool")
	}

	totalIPs := int64(0)
	for ip := dupIP(startIP); !ip.Equal(endIP); incrementIP(ip) {
		totalIPs++
	}
	totalIPs++ // Include the end IP

	// Get allocated count
	var allocatedCount int64
	if err := r.db.WithContext(ctx).Model(&model.IPAllocation{}).
		Where("ip_pool_id = ? AND status != ?", poolID, "available").
		Count(&allocatedCount).Error; err != nil {
		return 0, err
	}

	return totalIPs - allocatedCount, nil
}

// dupIP creates a copy of an IP address.
func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

// incrementIP increments an IP address by 1.
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
