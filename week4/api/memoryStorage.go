package main

import (
	"context"
	"strings"
	"strconv"
	"sync"
	"time"
	"fmt"
)


// map
type MemoryStorage struct {
	mu        sync.RWMutex //to avoid multiple http requests accessing map
	instances map[string]DBInstance
	nextID    int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		instances: make(map[string]DBInstance),
		nextID:    1,
	}
}

// list all instances for GET/instances
func (s *MemoryStorage) List(ctx context.Context) ([]DBInstance, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock() //unlock before function finish

	// Convert map -> slice
	instanceList := make([]DBInstance, 0, len(s.instances))

	// ignore id
	for _, instance := range s.instances {
		instanceList = append(instanceList, instance)
	}

	return instanceList, nil
}

// Get /instances/{id}
func (s *MemoryStorage) Get(ctx context.Context, id string) (DBInstance, error) {
	if err := ctx.Err(); err != nil {
		return DBInstance{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	instance, exist := s.instances[id]
	if !exist {
		return DBInstance{}, ErrInstanceNotFound
	}

	return instance, nil
}

// POST/instances
func (s *MemoryStorage) Create(ctx context.Context, r CreateInstanceRequest) (DBInstance, error) {

	if err := ctx.Err(); err != nil {
		return DBInstance{}, err
	}

	// prepare id and create instance
	id := strconv.Itoa(s.nextID)
	s.nextID++

	instance := DBInstance{
		ID:        id,
		Name:      r.Name,
		Instances: r.Instances,
		Status:    "pending",
		CreatedAt: time.Now().Format("2006-01-02T15:04:05"),
	}

	//save to map and return status
	s.instances[id] = instance
	return instance, nil
}

// PATCH /instances/{id}
func (s *MemoryStorage) Update(ctx context.Context, id string, r PatchInstanceRequest) (DBInstance, error) {
	if err := ctx.Err(); err != nil {
		return DBInstance{}, err
	}

	instance, exists := s.instances[id]
	if !exists {
		return DBInstance{}, ErrInstanceNotFound
	}

	//update
	instance.Name = r.Name
	instance.Instances = r.Instances

	s.instances[id] = instance
	return instance, nil
}

// DELETE /instances/{id}
func (s *MemoryStorage) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, exist := s.instances[id]
	if !exist {
		return ErrInstanceNotFound
	}

	delete(s.instances, id) // go function for delet map element
	return nil
}
