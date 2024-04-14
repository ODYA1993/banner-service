package banner

import (
	"context"
	"sync"
	"time"
)

type CacheBanner struct {
	cache map[string]*cacheItem
	mutex sync.RWMutex
	ttl   time.Duration
	quit  chan struct{}
}

type cacheItem struct {
	banner     *Banner
	expiration time.Time
}

func NewBannerCache(ttl time.Duration) *CacheBanner {
	cache := &CacheBanner{
		cache: make(map[string]*cacheItem),
		ttl:   ttl,
		quit:  make(chan struct{}),
	}
	go cache.runExpirationLoop()
	return cache
}

func (bc *CacheBanner) SetBanner(ctx context.Context, key string, banner *Banner) {
	expiration := time.Now().Add(bc.ttl)
	bc.mutex.Lock()
	bc.cache[key] = &cacheItem{
		banner:     banner,
		expiration: expiration,
	}
	bc.mutex.Unlock()
}

func (bc *CacheBanner) GetBanner(ctx context.Context, key string) (*Banner, bool) {
	bc.mutex.RLock()
	item, ok := bc.cache[key]
	bc.mutex.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(item.expiration) {
		bc.DeleteBanner(ctx, key)
		return nil, false
	}
	return item.banner, true
}

func (bc *CacheBanner) DeleteBanner(ctx context.Context, key string) {
	bc.mutex.Lock()
	delete(bc.cache, key)
	bc.mutex.Unlock()
}

func (bc *CacheBanner) runExpirationLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			bc.expireCache()
		case <-bc.quit:
			return
		}
	}
}

func (bc *CacheBanner) expireCache() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	now := time.Now()
	for key, item := range bc.cache {
		if now.After(item.expiration) {
			delete(bc.cache, key)
		}
	}
}

func (bc *CacheBanner) Close() {
	close(bc.quit)
}
