// Package service provides business logic implementations.
package service

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/Veritas-Calculus/vc-lab-platform/internal/model"
	"github.com/Veritas-Calculus/vc-lab-platform/internal/repository"
	"go.uber.org/zap"
)

// IPAMService defines the interface for IP Address Management operations.
type IPAMService interface {
	// Pool operations
	ListPools(ctx context.Context, zoneID string, page, pageSize int) ([]*model.IPPool, int64, error)
	GetPool(ctx context.Context, id string) (*model.IPPool, error)
	CreatePool(ctx context.Context, input *CreateIPPoolInput) (*model.IPPool, error)
	UpdatePool(ctx context.Context, id string, input *UpdateIPPoolInput) (*model.IPPool, error)
	DeletePool(ctx context.Context, id string) error

	// Allocation operations
	ListAllocations(ctx context.Context, poolID string, page, pageSize int) ([]*model.IPAllocation, int64, error)
	GetAllocation(ctx context.Context, id string) (*model.IPAllocation, error)
	AllocateIP(ctx context.Context, input *AllocateIPInput) (*model.IPAllocation, error)
	ReleaseIP(ctx context.Context, id string) error
	GetAllocationsByResource(ctx context.Context, resourceID string) ([]*model.IPAllocation, error)
	GetAvailableCount(ctx context.Context, poolID string) (int64, error)
}

// CreateIPPoolInput represents input for creating an IP pool.
type CreateIPPoolInput struct {
	Name        string
	CIDR        string
	Gateway     string
	DNS         string
	VLANTag     int
	StartIP     string
	EndIP       string
	ZoneID      string
	NetworkType string
	Description string
}

// UpdateIPPoolInput represents input for updating an IP pool.
type UpdateIPPoolInput struct {
	Name        *string
	Gateway     *string
	DNS         *string
	VLANTag     *int
	Description *string
	Status      *int8
}

// AllocateIPInput represents input for allocating an IP address.
type AllocateIPInput struct {
	PoolID     string
	Hostname   string
	ResourceID string
	IPAddress  string // Optional: specific IP to allocate, empty for next available
}

type ipamService struct {
	poolRepo       repository.IPPoolRepository
	allocationRepo repository.IPAllocationRepository
	logger         *zap.Logger
}

// NewIPAMService creates a new IPAM service.
func NewIPAMService(
	poolRepo repository.IPPoolRepository,
	allocationRepo repository.IPAllocationRepository,
	logger *zap.Logger,
) IPAMService {
	return &ipamService{
		poolRepo:       poolRepo,
		allocationRepo: allocationRepo,
		logger:         logger,
	}
}

// ListPools retrieves IP pools with pagination.
func (s *ipamService) ListPools(ctx context.Context, zoneID string, page, pageSize int) ([]*model.IPPool, int64, error) {
	offset := (page - 1) * pageSize
	return s.poolRepo.List(ctx, zoneID, offset, pageSize)
}

// GetPool retrieves an IP pool by ID.
func (s *ipamService) GetPool(ctx context.Context, id string) (*model.IPPool, error) {
	return s.poolRepo.GetByID(ctx, id)
}

// CreatePool creates a new IP pool.
func (s *ipamService) CreatePool(ctx context.Context, input *CreateIPPoolInput) (*model.IPPool, error) {
	// Validate CIDR
	_, ipNet, err := net.ParseCIDR(input.CIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	// Validate start and end IP
	startIP := net.ParseIP(input.StartIP)
	if startIP == nil {
		return nil, errors.New("invalid start IP address")
	}

	endIP := net.ParseIP(input.EndIP)
	if endIP == nil {
		return nil, errors.New("invalid end IP address")
	}

	// Check if IPs are within CIDR range
	if !ipNet.Contains(startIP) {
		return nil, errors.New("start IP is not within CIDR range")
	}
	if !ipNet.Contains(endIP) {
		return nil, errors.New("end IP is not within CIDR range")
	}

	// Validate gateway
	gateway := net.ParseIP(input.Gateway)
	if gateway == nil {
		return nil, errors.New("invalid gateway IP address")
	}

	pool := &model.IPPool{
		Name:        input.Name,
		CIDR:        input.CIDR,
		Gateway:     input.Gateway,
		DNS:         input.DNS,
		VLANTag:     input.VLANTag,
		StartIP:     input.StartIP,
		EndIP:       input.EndIP,
		ZoneID:      input.ZoneID,
		NetworkType: input.NetworkType,
		Description: input.Description,
		Status:      1, // 1: active
	}

	if err := s.poolRepo.Create(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to create IP pool: %w", err)
	}

	return pool, nil
}

// UpdatePool updates an existing IP pool.
func (s *ipamService) UpdatePool(ctx context.Context, id string, input *UpdateIPPoolInput) (*model.IPPool, error) {
	pool, err := s.poolRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		pool.Name = *input.Name
	}
	if input.Gateway != nil {
		gateway := net.ParseIP(*input.Gateway)
		if gateway == nil {
			return nil, errors.New("invalid gateway IP address")
		}
		pool.Gateway = *input.Gateway
	}
	if input.DNS != nil {
		pool.DNS = *input.DNS
	}
	if input.VLANTag != nil {
		pool.VLANTag = *input.VLANTag
	}
	if input.Description != nil {
		pool.Description = *input.Description
	}
	if input.Status != nil {
		pool.Status = *input.Status
	}

	if err := s.poolRepo.Update(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to update IP pool: %w", err)
	}

	return pool, nil
}

// DeletePool deletes an IP pool.
func (s *ipamService) DeletePool(ctx context.Context, id string) error {
	return s.poolRepo.Delete(ctx, id)
}

// ListAllocations retrieves IP allocations for a pool.
func (s *ipamService) ListAllocations(ctx context.Context, poolID string, page, pageSize int) ([]*model.IPAllocation, int64, error) {
	offset := (page - 1) * pageSize
	return s.allocationRepo.ListByPool(ctx, poolID, offset, pageSize)
}

// GetAllocation retrieves an IP allocation by ID.
func (s *ipamService) GetAllocation(ctx context.Context, id string) (*model.IPAllocation, error) {
	return s.allocationRepo.GetByID(ctx, id)
}

// AllocateIP allocates an IP address from a pool.
//
//nolint:nestif // nested conditions needed for IP validation and allocation logic
func (s *ipamService) AllocateIP(ctx context.Context, input *AllocateIPInput) (*model.IPAllocation, error) {
	if input.IPAddress != "" {
		// Check if the specific IP is available
		existing, err := s.allocationRepo.GetByIPAddress(ctx, input.PoolID, input.IPAddress)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		if existing != nil && existing.Status != "available" {
			return nil, errors.New("IP address is already allocated")
		}

		// Validate the IP is within the pool range
		pool, err := s.poolRepo.GetByID(ctx, input.PoolID)
		if err != nil {
			return nil, err
		}

		ip := net.ParseIP(input.IPAddress)
		if ip == nil {
			return nil, errors.New("invalid IP address")
		}

		startIP := net.ParseIP(pool.StartIP)
		endIP := net.ParseIP(pool.EndIP)
		if !isIPInRange(ip, startIP, endIP) {
			return nil, errors.New("IP address is not within pool range")
		}

		// Allocate the specific IP
		var resID *string
		if input.ResourceID != "" {
			resID = &input.ResourceID
		}
		allocation := &model.IPAllocation{
			IPPoolID:   input.PoolID,
			IPAddress:  input.IPAddress,
			Hostname:   input.Hostname,
			ResourceID: resID,
			Status:     "allocated",
		}

		if err := s.allocationRepo.Create(ctx, allocation); err != nil {
			return nil, fmt.Errorf("failed to allocate IP: %w", err)
		}

		return allocation, nil
	}

	// Allocate next available IP
	return s.allocationRepo.AllocateNextAvailable(ctx, input.PoolID, input.Hostname, input.ResourceID)
}

// ReleaseIP releases an allocated IP address.
func (s *ipamService) ReleaseIP(ctx context.Context, id string) error {
	return s.allocationRepo.Release(ctx, id)
}

// GetAllocationsByResource retrieves all IP allocations for a resource.
func (s *ipamService) GetAllocationsByResource(ctx context.Context, resourceID string) ([]*model.IPAllocation, error) {
	return s.allocationRepo.ListByResource(ctx, resourceID)
}

// GetAvailableCount returns the count of available IPs in a pool.
func (s *ipamService) GetAvailableCount(ctx context.Context, poolID string) (int64, error) {
	return s.allocationRepo.GetAvailableCount(ctx, poolID)
}

// isIPInRange checks if an IP is within the given range.
func isIPInRange(ip, start, end net.IP) bool {
	ip = ip.To16()
	start = start.To16()
	end = end.To16()

	for i := 0; i < len(ip); i++ {
		if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
	}
	return true
}
