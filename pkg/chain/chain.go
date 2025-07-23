package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"rk-api/pkg/logger"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// 智能限流器
type RateLimiter struct {
	limitChan chan struct{}
	ticker    *time.Ticker
}

func NewRateLimiter(rps int) *RateLimiter {
	rl := &RateLimiter{
		limitChan: make(chan struct{}, rps),
		ticker:    time.NewTicker(time.Second / time.Duration(rps)),
	}

	// 定时填充令牌
	go func() {
		for range rl.ticker.C {
			select {
			case rl.limitChan <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.limitChan:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Wait() {
	<-rl.limitChan
}

// 区块数据结构
type Block struct {
	Hash       string `json:"hash"`
	Number     uint64 `json:"number"`
	Timestamp  int64  `json:"timestamp"`
	ParentHash string `json:"parentHash"`
}

// 区块获取器
type BlockFetcher struct {
	apiEndpoints []string // 多个API端点
	cache        *BlockCache
	rateLimiter  *RateLimiter
	httpClient   *http.Client
	predictCache *PredictionCache

	currentHeight uint64        // 原子操作的当前高度
	heightSubs    []chan uint64 // 高度订阅通道
	subsMu        sync.RWMutex  // 订阅锁
}

// 初始化区块获取器
func NewBlockFetcher(endpoints []string, rps int) *BlockFetcher {
	return &BlockFetcher{
		apiEndpoints: endpoints,
		cache:        NewBlockCache(20),   // 缓存最近20个区块
		rateLimiter:  NewRateLimiter(rps), //因为 api 限制了请求频率，所以这里设置一个限流器
		httpClient:   &http.Client{Timeout: 3 * time.Second},
		predictCache: NewPredictionCache(),
	}
}

// 订阅区块高度变化
func (f *BlockFetcher) SubscribeHeight() <-chan uint64 {
	f.subsMu.Lock()
	defer f.subsMu.Unlock()

	ch := make(chan uint64, 20) // 带缓冲的通道
	f.heightSubs = append(f.heightSubs, ch)
	return ch
}

// 获取最新区块高度（无API调用）
func (f *BlockFetcher) GetLatestHeight() uint64 {
	return atomic.LoadUint64(&f.currentHeight)
}

// updateLatestBlock
func (f *BlockFetcher) updateLatestBlock(block *Block) {
	f.cache.Add(block)

	// 更新当前高度并通知订阅者
	old := atomic.SwapUint64(&f.currentHeight, block.Number)
	if old != block.Number {
		f.notifySubscribers(block.Number)
	}
}

// 通知所有订阅者
func (f *BlockFetcher) notifySubscribers(height uint64) {
	f.subsMu.RLock()
	defer f.subsMu.RUnlock()

	for _, ch := range f.heightSubs {
		select {
		case ch <- height:
		default: // 防止慢消费者阻塞
		}
	}
}

// 获取指定高度区块（带缓存）
func (f *BlockFetcher) GetBlock(height uint64) (*Block, error) {
	// 优先从缓存获取
	if block := f.cache.Get(height); block != nil {
		logger.Info("GetBlock from cache", zap.Uint64("height", height))
		return block, nil
	}

	// 控制请求频率
	if !f.rateLimiter.Allow() {
		logger.Warn("GetBlock Rate limit exceeded")
		return nil, fmt.Errorf("rate limit exceeded")
	}

	logger.Info("GetBlock from API", zap.Uint64("height", height))

	// 轮询多个API端点
	var lastErr error
	for _, endpoint := range f.apiEndpoints {
		url := fmt.Sprintf("%s/block?number=%d", endpoint, height)
		block, err := f.fetchBlock(url)
		if err == nil {
			f.cache.Add(block)
			return block, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// 获取最新区块（特殊处理）
func (f *BlockFetcher) GetLatestBlock() (*Block, error) {
	latesttime, subtimems, block := f.cache.GetLatesttime()
	if latesttime != 0 && time.Now().UnixMilli() < latesttime+subtimems {
		// logger.Info("GetLatestBlock time limit, from cache", zap.Uint64("height", block.Number))
		return block, nil
	}

	f.rateLimiter.Wait()

	var lastErr error
	for _, endpoint := range f.apiEndpoints {
		url := endpoint + "/block/latest"
		block, err := f.fetchLatestBlock(url)
		if err == nil {
			return block, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (f *BlockFetcher) GetCachedLatestBlockHash() string {
	block := f.cache.GetLatest()
	return block.Hash
}

// 定时更新最新区块
func (f *BlockFetcher) StartBackgroundUpdate(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if block, err := f.GetLatestBlock(); err == nil {
				f.updateLatestBlock(block)
			}
		case <-ctx.Done():
			return
		}
	}
}

// 私有方法
func (f *BlockFetcher) fetchBlock(url string) (*Block, error) {
	data, err := f.fetchBlockInner(url)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Block `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("block not found for url: %s, data: %s", url, string(data))
	}

	return &result.Data[0], nil
}

// 私有方法
func (f *BlockFetcher) fetchLatestBlock(url string) (*Block, error) {
	data, err := f.fetchBlockInner(url)
	if err != nil {
		return nil, err
	}
	var result *Block
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("block not found for url: %s, data: %s", url, string(data))
	}
	return result, nil
}

func (f *BlockFetcher) fetchBlockInner(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// 缓存实现
type BlockCache struct {
	latesttime int64
	subtimems  int64
	block      *Block
	blocks     map[uint64]*Block
	mu         sync.RWMutex
	size       int
}

func NewBlockCache(size int) *BlockCache {
	return &BlockCache{
		blocks: make(map[uint64]*Block),
		size:   size,
	}
}

func (c *BlockCache) Add(block *Block) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.blocks) >= c.size {
		// 简单LRU策略
		var minKey uint64
		for k := range c.blocks {
			if minKey == 0 || k < minKey {
				minKey = k
			}
		}
		delete(c.blocks, minKey)
	}
	c.blocks[block.Number] = block
	c.block = block
	c.latesttime = block.Timestamp
	c.subtimems = time.Now().UnixMilli() - block.Timestamp
}

func (c *BlockCache) Get(height uint64) *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blocks[height]
}

func (c *BlockCache) GetLatest() *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.block
}

func (c *BlockCache) GetLatesttime() (int64, int64, *Block) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latesttime, c.subtimems, c.block
}

// 预测缓存
type PredictionCache struct {
	current   *Block
	subtimems int64
	predicted *Block
	mu        sync.RWMutex
}

func NewPredictionCache() *PredictionCache {
	return &PredictionCache{}
}

func (c *PredictionCache) Update(block *Block) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		c.current = block
		c.subtimems = time.Now().UnixMilli() - block.Timestamp
		c.predicted = &Block{
			Number:    block.Number + 1,
			Timestamp: block.Timestamp + 3000, // 假设3秒出块
		}
		// logger.Info("updatePrediction init", zap.Uint64("number", block.Number), zap.Int64("timestamp", block.Timestamp))
	} else if time.Now().UnixMilli() >= block.Timestamp+c.subtimems {
		c.current = block
		c.predicted = &Block{
			Number:    block.Number + 1,
			Timestamp: block.Timestamp + 3000, // 假设3秒出块
		}
		// logger.Info("updatePrediction current", zap.Uint64("number", block.Number), zap.Int64("timestamp", block.Timestamp))
	}
}

func (c *PredictionCache) GetLatest() *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.predicted
}

// fetcher := NewBlockFetcher(endpoints, 5) // 5 requests per second
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	go fetcher.StartBackgroundUpdate(ctx)

// 	// 游戏房间使用示例
// 	for {
// 		current, err := fetcher.GetLatestBlock()
// 		if err != nil {
// 			// 处理错误
// 			continue
// 		}

// 		nextBlock := fetcher.predictCache.GetLatest()
// 		fmt.Printf("当前区块: %d 预测下个区块时间: %d\n",
// 			current.Number, nextBlock.Timestamp)

// 		time.Sleep(1 * time.Second)
// 	}
