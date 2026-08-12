package instance

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
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

	// prepare params and validate
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return DBInstance{}, fmt.Errorf(
			"%w: name is required",
			ErrInvalidInstance,
		)
	}

	if r.Instances < 1 {
		return DBInstance{}, fmt.Errorf(
			"%w: instances must be at least 1",
			ErrInvalidInstance,
		)
	}

	storage := defaultStorageSize
	if r.Storage != nil {
		value, err := normalizePositiveQuantity(*r.Storage, "storage")
		if err != nil {
			return DBInstance{}, err
		}
		storage = value
	}

	cpu := defaultCPURequest
	if r.CPU != nil {
		value, err := normalizePositiveQuantity(*r.CPU, "cpu")
		if err != nil {
			return DBInstance{}, err
		}
		cpu = value
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := strconv.Itoa(s.nextID)
	s.nextID++

	instance := DBInstance{
		ID:        id,
		Name:      name,
		Instances: r.Instances,
		Storage:   storage,
		CPU:       cpu,
		Status:    "pending",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	//save to map and return status
	s.instances[id] = instance
	return instance, nil
}

// PATCH /instances/{id}
func (s *MemoryStorage) Patch(ctx context.Context, id string, r PatchInstanceRequest) (DBInstance, error) {
	if err := ctx.Err(); err != nil {
		return DBInstance{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	instance, exists := s.instances[id]
	if !exists {
		return DBInstance{}, ErrInstanceNotFound
	}

	if r.Name != nil {
		name := strings.TrimSpace(*r.Name)
		if name == "" {
			return DBInstance{}, fmt.Errorf(
				"%w: name cannot be empty",
				ErrInvalidInstance,
			)
		}
		instance.Name = name
	}

	if r.Instances != nil {
		if *r.Instances < 1 {
			return DBInstance{}, fmt.Errorf(
				"%w: instances must be at least 1",
				ErrInvalidInstance,
			)
		}
		instance.Instances = *r.Instances
	}

	if r.Storage != nil {
		storage, err := normalizePositiveQuantity(*r.Storage, "storage")
		if err != nil {
			return DBInstance{}, err
		}

		currentStorage := instance.Storage
		if currentStorage == "" {
			currentStorage = defaultStorageSize
		}

		if err := validateStorageExpansion(currentStorage, storage); err != nil {
			return DBInstance{}, err
		}

		instance.Storage = storage
	}

	if r.CPU != nil {
		cpu, err := normalizePositiveQuantity(*r.CPU, "cpu")
		if err != nil {
			return DBInstance{}, err
		}
		instance.CPU = cpu
	}

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
