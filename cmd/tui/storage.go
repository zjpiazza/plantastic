package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/zjpiazza/plantastic/internal/models"
)

// Storage defines the interface for interacting with data
type Storage interface {
	// Garden methods
	GetGardens() []models.Garden
	GetGarden(id string) (models.Garden, bool)
	AddGarden(garden models.Garden) error
	UpdateGarden(garden models.Garden) error
	DeleteGarden(id string) error

	// Bed methods
	GetBeds(gardenID string) []models.Bed
	GetBed(id string) (models.Bed, bool)
	AddBed(bed models.Bed) error
	UpdateBed(bed models.Bed) error
	DeleteBed(id string) error

	// Task methods
	GetTasks(gardenID string, bedID *string) []models.Task
	GetTask(id string) (models.Task, bool)
	AddTask(task models.Task) error
	UpdateTask(task models.Task) error
	DeleteTask(id string) error
}

// MemoryStorage provides in-memory storage for gardens, beds, and tasks
type MemoryStorage struct {
	gardens map[string]models.Garden
	beds    map[string]models.Bed
	tasks   map[string]models.Task
	mu      sync.RWMutex
}

// NewMemoryStorage creates a new instance of MemoryStorage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		gardens: make(map[string]models.Garden),
		beds:    make(map[string]models.Bed),
		tasks:   make(map[string]models.Task),
	}
}

// Garden operations
func (s *MemoryStorage) GetGarden(id string) (models.Garden, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	garden, ok := s.gardens[id]
	return garden, ok
}

func (s *MemoryStorage) GetGardens() []models.Garden {
	s.mu.RLock()
	defer s.mu.RUnlock()
	gardens := make([]models.Garden, 0, len(s.gardens))
	for _, garden := range s.gardens {
		gardens = append(gardens, garden)
	}
	return gardens
}

func (s *MemoryStorage) AddGarden(garden models.Garden) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.gardens[garden.ID]; exists {
		return fmt.Errorf("garden with ID %s already exists", garden.ID)
	}
	s.gardens[garden.ID] = garden
	return nil
}

func (s *MemoryStorage) UpdateGarden(garden models.Garden) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.gardens[garden.ID]; !exists {
		return fmt.Errorf("garden with ID %s not found", garden.ID)
	}
	garden.UpdatedAt = time.Now()
	s.gardens[garden.ID] = garden
	return nil
}

func (s *MemoryStorage) DeleteGarden(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.gardens[id]; !exists {
		return fmt.Errorf("garden with ID %s not found", id)
	}
	delete(s.gardens, id)
	return nil
}

// Bed operations
func (s *MemoryStorage) GetBed(id string) (models.Bed, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bed, ok := s.beds[id]
	return bed, ok
}

func (s *MemoryStorage) GetBeds(gardenID string) []models.Bed {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var beds []models.Bed
	for _, bed := range s.beds {
		if bed.GardenID == gardenID {
			beds = append(beds, bed)
		}
	}
	return beds
}

func (s *MemoryStorage) AddBed(bed models.Bed) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.beds[bed.ID]; exists {
		return fmt.Errorf("bed with ID %s already exists", bed.ID)
	}
	s.beds[bed.ID] = bed
	return nil
}

func (s *MemoryStorage) UpdateBed(bed models.Bed) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.beds[bed.ID]; !exists {
		return fmt.Errorf("bed with ID %s not found", bed.ID)
	}
	bed.UpdatedAt = time.Now()
	s.beds[bed.ID] = bed
	return nil
}

func (s *MemoryStorage) DeleteBed(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.beds[id]; !exists {
		return fmt.Errorf("bed with ID %s not found", id)
	}
	delete(s.beds, id)
	return nil
}

// Task operations
func (s *MemoryStorage) GetTask(id string) (models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	return task, ok
}

func (s *MemoryStorage) GetTasks(gardenID string, bedID *string) []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tasks []models.Task
	for _, task := range s.tasks {
		if task.GardenID == gardenID {
			if bedID == nil || (task.BedID != nil && *task.BedID == *bedID) {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}

func (s *MemoryStorage) AddTask(task models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStorage) UpdateTask(task models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[task.ID]; !exists {
		return fmt.Errorf("task with ID %s not found", task.ID)
	}
	task.UpdatedAt = time.Now()
	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStorage) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[id]; !exists {
		return fmt.Errorf("task with ID %s not found", id)
	}
	delete(s.tasks, id)
	return nil
}
