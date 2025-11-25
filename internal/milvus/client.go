package milvus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"go.uber.org/zap"
)

// Config Milvus 配置
type Config struct {
	Host           string        // Milvus 服务器地址
	Port           int           // 端口号
	Username       string        // 用户名（可选）
	Password       string        // 密码（可选）
	Database       string        // 数据库名称（可选）
	ConnectTimeout time.Duration // 连接超时时间
	EnableTLS      bool          // 是否启用 TLS
	TLSCert        string        // TLS 证书路径
}

// Client Milvus 客户端封装
type Client struct {
	config       Config
	milvusClient client.Client
	logger       *zap.Logger
	mu           sync.RWMutex
	isConnected  bool
}

// NewClient 创建新的 Milvus 客户端
func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	// 构建连接参数
	connectParams := client.Config{
		Address: fmt.Sprintf("%s:%d", config.Host, config.Port),
	}

	if config.Username != "" {
		connectParams.Username = config.Username
		connectParams.Password = config.Password
	}

	if config.Database != "" {
		connectParams.DBName = config.Database
	}

	if config.EnableTLS {
		connectParams.EnableTLSAuth = true
	}

	// 创建 Milvus 客户端
	milvusClient, err := client.NewClient(context.Background(), connectParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus client: %w", err)
	}

	c := &Client{
		config:       config,
		milvusClient: milvusClient,
		logger:       logger,
		isConnected:  true,
	}

	// 测试连接
	if err := c.Ping(context.Background()); err != nil {
		milvusClient.Close()
		return nil, fmt.Errorf("failed to ping milvus: %w", err)
	}

	logger.Info("milvus client connected",
		zap.String("address", connectParams.Address),
	)

	return c, nil
}

// GetClient 获取底层 Milvus 客户端（用于高级操作）
func (c *Client) GetClient() client.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.milvusClient
}

// Ping 检查连接是否正常
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected {
		return fmt.Errorf("client is not connected")
	}

	// 尝试列出集合来测试连接
	_, err := c.milvusClient.ListCollections(ctx)
	if err != nil {
		c.logger.Error("ping milvus failed", zap.Error(err))
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.milvusClient != nil {
		if err := c.milvusClient.Close(); err != nil {
			c.logger.Error("close milvus client failed", zap.Error(err))
			return err
		}
	}
	c.isConnected = false
	c.logger.Info("milvus client closed")
	return nil
}

// IsConnected 返回连接状态
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}
