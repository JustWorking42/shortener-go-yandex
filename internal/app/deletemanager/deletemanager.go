package deletemanager

import (
	"context"
	"time"

	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"go.uber.org/zap"
)

type DeleteManager struct {
	Storage  storage.Storage
	TaskChan chan models.DeleteTask
	Logger   *zap.Logger
}

func NewDeleteManager(storage storage.Storage) *DeleteManager {
	return &DeleteManager{
		Storage:  storage,
		TaskChan: make(chan models.DeleteTask, 256),
	}
}

func (m *DeleteManager) SubcribeOnTask(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 5)
	var taskSlice []models.DeleteTask
	go func() {
		for {
			select {

			case task := <-m.TaskChan:
				taskSlice = append(taskSlice, task)

			case <-ticker.C:
				if len(taskSlice) > 0 {
					err := m.Storage.Delete(ctx, taskSlice)
					if err != nil {
						m.Logger.Sugar().Errorf("delete err numbers of elements: %d", len(taskSlice))
						continue
					}
					taskSlice = nil
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
