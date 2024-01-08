// Package deletemanager provides functionality for managing deletion tasks.
package deletemanager

import (
	"context"
	"sync"
	"time"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"go.uber.org/zap"
)

// DeleteManager is responsible for managing deletion tasks and processing them.
type DeleteManager struct {
	Storage  storage.Storage
	TaskChan chan models.DeleteTask
	Logger   *zap.Logger
}

// NewDeleteManager creates a new instance of DeleteManager with the provided storage.
func NewDeleteManager(storage storage.Storage) *DeleteManager {
	return &DeleteManager{
		Storage:  storage,
		TaskChan: make(chan models.DeleteTask, 256),
	}
}

// SubcribeOnTask starts the process of listening for deletion tasks and processing them.
func (m *DeleteManager) SubcribeOnTask(ctx context.Context) (*sync.WaitGroup, chan error) {
	ticker := time.NewTicker(time.Second * 5)
	var taskSlice []models.DeleteTask
	var wg sync.WaitGroup
	errChan := make(chan error)
	wg.Add(1)
	go func() {
		for {
			select {
			case task := <-m.TaskChan:
				taskSlice = append(taskSlice, task)

			case <-ticker.C:
				if len(taskSlice) > 0 {
					err := m.Storage.Delete(ctx, taskSlice)
					if err != nil {
						errChan <- err
						continue
					}
					taskSlice = nil
				}

			case <-ctx.Done():
				if len(taskSlice) > 0 {
					err := m.Storage.Delete(context.Background(), taskSlice)
					if err != nil {
						errChan <- err
					}
				}
				wg.Done()
				close(errChan)
				return
			}
		}
	}()
	return &wg, errChan
}
