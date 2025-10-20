package cache

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3" // 确保你已经 get 了这个包
)

// 确保 SqliteCache 实现了 Cache 接口
var _ Cache = (*SqliteCache)(nil)

// SqliteCache 使用 SQLite 数据库实现 Cache 接口。
type SqliteCache struct {
	db *sql.DB
}

// NewSqliteCache 创建一个新的 SqliteCache 实例。
// 它会打开指定路径的数据库文件，并创建缓存表（如果不存在）。
func NewSqliteCache(dbPath string) (*SqliteCache, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL") // 使用 WAL 模式提高并发性能
	if err != nil {
		return nil, err
	}

	// 创建缓存表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cache (
			key TEXT PRIMARY KEY,
			value TEXT,
			expires_at INTEGER
		)
	`)
	if err != nil {
		return nil, err
	}

	cache := &SqliteCache{db: db}

	// 启动一个 goroutine 定期清理过期的键
	go cache.cleanupExpiredKeys()

	return cache, nil
}

// Set 将一个键值对和 TTL 添加到缓存中。
func (c *SqliteCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).Unix()
	}
	// expiresAt 为 0 表示永不过期

	_, err := c.db.ExecContext(ctx, `
		REPLACE INTO cache (key, value, expires_at) VALUES (?, ?, ?)
	`, key, value, expiresAt)
	return err
}

// Get 通过键从缓存中检索值。
// 如果键不存在或已过期，将返回错误。
func (c *SqliteCache) Get(ctx context.Context, key string) (any, error) {
	var value string
	var expiresAt int64

	row := c.db.QueryRowContext(ctx, `
		SELECT value, expires_at FROM cache WHERE key = ?
	`, key)

	err := row.Scan(&value, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", newKeyNotExistsError(key)
		}
		return "", err // sql.ErrNoRows 是这里的常见错误
	}

	// 检查是否过期
	if expiresAt != 0 && time.Now().Unix() > expiresAt {
		// 键已过期，删除它并返回未找到
		_ = c.Del(ctx, key) // 尽力而为地删除
		return "", newKeyNotExistsError(key)
	}

	return value, nil
}

// Del 从缓存中删除一个键。
func (c *SqliteCache) Del(ctx context.Context, key string) error {
	_, err := c.db.ExecContext(ctx, `
		DELETE FROM cache WHERE key = ?
	`, key)
	return err
}

// Has 检查缓存中是否存在一个未过期的键。
func (c *SqliteCache) Has(ctx context.Context, key string) (bool, error) {
	var expiresAt int64

	row := c.db.QueryRowContext(ctx, `
		SELECT expires_at FROM cache WHERE key = ?
	`, key)

	err := row.Scan(&expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// 检查是否过期
	if expiresAt != 0 && time.Now().Unix() > expiresAt {
		_ = c.Del(ctx, key) // 尽力而为地删除
		return false, nil
	}

	return true, nil
}

// SetWithFunc 尝试获取一个键，如果不存在，则调用函数，
// 设置缓存，并返回结果。
func (c *SqliteCache) SetWithFunc(ctx context.Context, key string, fn func() (any, error), ttl time.Duration) (any, error) {
	// 缓存中没有，调用函数
	val, err := fn()
	if err != nil {
		return "", err
	}

	// 在缓存中设置新值。这里我们使用一个默认的 TTL，你可能希望将其作为参数传入。
	// 在这个例子中，我们假设默认 TTL 为 10 分钟。
	err = c.Set(ctx, key, val, ttl)
	if err != nil {
		// 记录错误但仍然返回值
		// log.Printf("failed to set cache for key %s: %v", key, err)
		return "", err
	}

	return val, nil
}

// cleanupExpiredKeys 定期从数据库中删除过期的键。
func (c *SqliteCache) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Unix()
		_, _ = c.db.Exec(`DELETE FROM cache WHERE expires_at != 0 AND expires_at <= ?`, now)
	}
}

// Close 关闭数据库连接。
func (c *SqliteCache) Close() error {
	return c.db.Close()
}
