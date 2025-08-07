// Package workers содержит фоновые рабочие горутины для обработки задач удаления URL.
package workers

import (
	"context"

	"github.com/aseptimu/url-shortener/internal/app/handlers/http/shortenurlhandlers"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"go.uber.org/zap"
)

// StartDeleteWorkerPool запускает пул из numWorkers воркеров, каждый из которых:
//  1. слушает контекст ctx на завершение работы;
//  2. читает задачи удаления URL из канала shortenurlhandlers.DeleteTaskCh;
//  3. при получении задачи вызывает deleter.DeleteURLs для пакетного удаления;
//  4. логирует успешное или ошибочное выполнение.
//
// workerID используется в логах для идентификации конкретного воркера.
func StartDeleteWorkerPool(ctx context.Context, numWorkers int, deleter service.URLDeleter, logger *zap.SugaredLogger) {
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			logger.Infow("Delete worker started", "workerID", workerID)
			for {
				select {
				case <-ctx.Done():
					logger.Infow("Delete worker stopping", "workerID", workerID)
					return
				case task := <-shortenurlhandlers.DeleteTaskCh:
					if err := deleter.DeleteURLs(context.Background(), task.URLs, task.UserID); err != nil {
						logger.Errorw("Worker failed to delete URLs", "workerID", workerID, "error", err)
					} else {
						logger.Infow("Worker deleted URLs successfully", "workerID", workerID)
					}
				}
			}
		}(i)
	}
}
