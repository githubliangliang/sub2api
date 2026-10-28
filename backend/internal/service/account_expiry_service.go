package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// AccountExpiryService periodically pauses expired accounts when auto-pause is enabled.
type AccountExpiryService struct {
	accountRepo AccountRepository
	interval    time.Duration
	workerCtx   context.Context
	workerStop  context.CancelFunc
	lifecycleMu sync.Mutex
	startOnce   sync.Once
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

func NewAccountExpiryService(accountRepo AccountRepository, interval time.Duration) *AccountExpiryService {
	ctx, cancel := context.WithCancel(context.Background())
	return &AccountExpiryService{
		accountRepo: accountRepo,
		interval:    interval,
		workerCtx:   ctx,
		workerStop:  cancel,
	}
}

func (s *AccountExpiryService) Start() {
	if s == nil || s.accountRepo == nil || s.interval <= 0 {
		return
	}
	s.startOnce.Do(func() {
		s.lifecycleMu.Lock()
		defer s.lifecycleMu.Unlock()
		if s.workerCtx.Err() != nil {
			return
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(s.interval)
			defer ticker.Stop()

			s.runOnce()
			for {
				select {
				case <-ticker.C:
					s.runOnce()
				case <-s.workerCtx.Done():
					return
				}
			}
		}()
	})
}

func (s *AccountExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		s.lifecycleMu.Lock()
		s.workerStop()
		s.lifecycleMu.Unlock()
	})
	s.wg.Wait()
}

func (s *AccountExpiryService) runOnce() {
	if s.workerCtx.Err() != nil {
		return
	}
	ctx, cancel := context.WithTimeout(s.workerCtx, 5*time.Second)
	defer cancel()

	updated, err := s.accountRepo.AutoPauseExpiredAccounts(ctx, time.Now())
	if err != nil {
		log.Printf("[AccountExpiry] Auto pause expired accounts failed: %v", err)
		return
	}
	if updated > 0 {
		log.Printf("[AccountExpiry] Auto paused %d expired accounts", updated)
	}
}
