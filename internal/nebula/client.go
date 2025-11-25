package nebula

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	nebula_go "github.com/vesoft-inc/nebula-go/v3"
	"go.uber.org/zap"
)

// Config Nebula Graph 配置
type Config struct {
	Addresses []string      // GraphD 地址列表，如 ["127.0.0.1:9669"]
	Username  string        // 用户名
	Password  string        // 密码
	Space     string        // 图空间名称
	Timeout   time.Duration // 连接超时时间
	MaxConn   int           // 最大连接数
	MinConn   int           // 最小连接数
}

// Client Nebula Graph 客户端封装
type Client struct {
	config      Config
	pool        *nebula_go.ConnectionPool
	session     *nebula_go.Session
	logger      *zap.Logger
	mu          sync.RWMutex
	isConnected bool
}

// NewClient 创建新的 Nebula Graph 客户端
func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	poolConfig := nebula_go.PoolConfig{
		TimeOut:         config.Timeout,
		IdleTime:        0,
		MaxConnPoolSize: config.MaxConn,
		MinConnPoolSize: config.MinConn,
	}

	addresses, err := parseAddresses(config.Addresses)
	if err != nil {
		return nil, fmt.Errorf("invalid nebula addresses: %w", err)
	}

	pool, err := nebula_go.NewConnectionPool(addresses, poolConfig, newNebulaLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 创建会话
	session, err := pool.GetSession(config.Username, config.Password)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 使用指定的图空间
	if config.Space != "" {
		_, err = session.Execute(fmt.Sprintf("USE %s", config.Space))
		if err != nil {
			session.Release()
			pool.Close()
			return nil, fmt.Errorf("failed to use space %s: %w", config.Space, err)
		}
	}

	c := &Client{
		config:      config,
		pool:        pool,
		session:     session,
		logger:      logger,
		isConnected: true,
	}

	return c, nil
}

// Execute 执行 nGQL 查询
func (c *Client) Execute(ctx context.Context, stmt string) (*nebula_go.ResultSet, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected {
		return nil, fmt.Errorf("client is not connected")
	}

	start := time.Now()
	resultSet, err := c.session.Execute(stmt)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("execute nGQL failed",
			zap.String("stmt", stmt),
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		return nil, fmt.Errorf("execute nGQL failed: %w", err)
	}

	if !resultSet.IsSucceed() {
		errMsg := resultSet.GetErrorMsg()
		c.logger.Error("nGQL execution failed",
			zap.String("stmt", stmt),
			zap.String("error", errMsg),
			zap.Duration("duration", duration),
		)
		return nil, fmt.Errorf("nGQL execution failed: %s", errMsg)
	}

	c.logger.Debug("execute nGQL success",
		zap.String("stmt", stmt),
		zap.Duration("duration", duration),
	)

	return resultSet, nil
}

// ExecuteWithRetry 带重试的执行
func (c *Client) ExecuteWithRetry(ctx context.Context, stmt string, maxRetries int) (*nebula_go.ResultSet, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resultSet, err := c.Execute(ctx, stmt)
		if err == nil {
			return resultSet, nil
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}
	return nil, fmt.Errorf("execute with retry failed after %d attempts: %w", maxRetries, lastErr)
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session != nil {
		c.session.Release()
	}
	if c.pool != nil {
		c.pool.Close()
	}
	c.isConnected = false
	c.logger.Info("nebula client closed")
	return nil
}

// Ping 检查连接是否正常
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Execute(ctx, "SHOW SPACES")
	return err
}

// IsConnected 返回连接状态
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// parseAddresses converts host:port strings to Nebula HostAddress list.
func parseAddresses(addresses []string) ([]nebula_go.HostAddress, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses provided")
	}

	result := make([]nebula_go.HostAddress, 0, len(addresses))
	for _, addr := range addresses {
		host, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("split host port for %s: %w", addr, err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("parse port for %s: %w", addr, err)
		}
		result = append(result, nebula_go.HostAddress{
			Host: host,
			Port: port,
		})
	}

	return result, nil
}

// zapNebulaLogger adapts zap.Logger to nebula_go.Logger interface.
type zapNebulaLogger struct {
	logger *zap.Logger
}

func newNebulaLogger(logger *zap.Logger) nebula_go.Logger {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &zapNebulaLogger{logger: logger}
}

func (l *zapNebulaLogger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *zapNebulaLogger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *zapNebulaLogger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *zapNebulaLogger) Fatal(msg string) {
	l.logger.Fatal(msg)
}
