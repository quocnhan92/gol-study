package main

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheItem đại diện cho một mục trong cache
type CacheItem struct {
	Data      interface{} // Dữ liệu cần lưu trữ trong cache
	ExpiredAt time.Time   // Thời điểm hết hạn của mục cache
}

// Cache là cấu trúc chứa dữ liệu cache và các phương thức liên quan
type Cache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
}

// NewCache tạo một cache mới
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

// Get lấy dữ liệu từ cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found || item.ExpiredAt.Before(time.Now()) {
		return nil, false
	}
	return item.Data, true
}

// Set đặt dữ liệu vào cache
func (c *Cache) Set(key string, data interface{}, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = CacheItem{
		Data:      data,
		ExpiredAt: time.Now().Add(duration),
	}
}

func main() {
	r := gin.Default()

	// Tạo một cache mới
	cache := NewCache()

	r.GET("/cached", func(c *gin.Context) {
		// Thử lấy dữ liệu từ cache
		if data, found := cache.Get("cached_data"); found {
			c.JSON(200, gin.H{"data": data})
			return
		}

		// Nếu không tìm thấy trong cache, thực hiện các công việc lấy dữ liệu
		// ở đây bạn có thể thực hiện các công việc như truy vấn cơ sở dữ liệu, tính toán phức tạp, v.v.

		// Giả sử ta có dữ liệu cần cache là "Hello, world!"
		data := "Hello, world!"

		// Đặt dữ liệu vào cache với thời gian hết hạn là 1 phút
		cache.Set("cached_data", data, time.Minute)

		c.JSON(200, gin.H{"data": data})
	})

	r.Run(":8080")
}
