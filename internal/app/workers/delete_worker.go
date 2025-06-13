package workers

import (
	"context"

	"github.com/aseptimu/url-shortener/internal/app/handlers/shortenurlhandlers"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"go.uber.org/zap"
)

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
