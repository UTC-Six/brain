package nebula

import (
	"context"
	"fmt"
	"strings"

	nebula_go "github.com/vesoft-inc/nebula-go/v3"
	"go.uber.org/zap"
)

// Vertex 顶点数据
type Vertex struct {
	VID   string                 // 顶点ID
	Tag   string                 // 标签名称
	Props map[string]interface{} // 属性
}

// VertexService 顶点操作服务
type VertexService struct {
	client *Client
	logger *zap.Logger
}

// NewVertexService 创建顶点服务
func NewVertexService(client *Client, logger *zap.Logger) *VertexService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &VertexService{
		client: client,
		logger: logger,
	}
}

// Create 创建顶点
// INSERT VERTEX tag_name (prop1, prop2, ...) VALUES vid:(value1, value2, ...)
func (s *VertexService) Create(ctx context.Context, vertex Vertex) error {
	if vertex.VID == "" || vertex.Tag == "" {
		return fmt.Errorf("VID and Tag are required")
	}

	var props []string
	var values []string
	for key, value := range vertex.Props {
		props = append(props, key)
		values = append(values, formatValue(value))
	}

	stmt := fmt.Sprintf(
		"INSERT VERTEX %s (%s) VALUES \"%s\":(%s)",
		vertex.Tag,
		strings.Join(props, ", "),
		vertex.VID,
		strings.Join(values, ", "),
	)

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("create vertex failed",
			zap.String("tag", vertex.Tag),
			zap.String("vid", vertex.VID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("vertex created",
		zap.String("tag", vertex.Tag),
		zap.String("vid", vertex.VID),
	)
	return nil
}

// BatchCreate 批量创建顶点
func (s *VertexService) BatchCreate(ctx context.Context, vertices []Vertex) error {
	if len(vertices) == 0 {
		return nil
	}

	// 按 Tag 分组
	tagGroups := make(map[string][]Vertex)
	for _, v := range vertices {
		tagGroups[v.Tag] = append(tagGroups[v.Tag], v)
	}

	// 为每个 Tag 批量插入
	for tag, vs := range tagGroups {
		if err := s.batchCreateByTag(ctx, tag, vs); err != nil {
			return err
		}
	}

	return nil
}

// batchCreateByTag 按标签批量创建
func (s *VertexService) batchCreateByTag(ctx context.Context, tag string, vertices []Vertex) error {
	if len(vertices) == 0 {
		return nil
	}

	// 获取第一个顶点的属性键作为模板
	firstVertex := vertices[0]
	var props []string
	for key := range firstVertex.Props {
		props = append(props, key)
	}

	// 构建批量插入语句
	var valueParts []string
	for _, v := range vertices {
		var values []string
		for _, prop := range props {
			value := firstVertex.Props[prop]
			if val, ok := v.Props[prop]; ok {
				value = val
			}
			values = append(values, formatValue(value))
		}
		valueParts = append(valueParts, fmt.Sprintf("\"%s\":(%s)", v.VID, strings.Join(values, ", ")))
	}

	stmt := fmt.Sprintf(
		"INSERT VERTEX %s (%s) VALUES %s",
		tag,
		strings.Join(props, ", "),
		strings.Join(valueParts, ", "),
	)

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("batch create vertices failed",
			zap.String("tag", tag),
			zap.Int("count", len(vertices)),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("vertices batch created",
		zap.String("tag", tag),
		zap.Int("count", len(vertices)),
	)
	return nil
}

// Get 查询顶点
// FETCH PROP ON tag_name vid
func (s *VertexService) Get(ctx context.Context, tag string, vid string) (*Vertex, error) {
	stmt := fmt.Sprintf("FETCH PROP ON %s \"%s\"", tag, vid)
	resultSet, err := s.client.Execute(ctx, stmt)
	if err != nil {
		return nil, err
	}

	if !resultSet.IsEmpty() {
		row, err := resultSet.GetRowValuesByIndex(0)
		if err != nil {
			return nil, err
		}

		vertex := &Vertex{
			VID:   vid,
			Tag:   tag,
			Props: make(map[string]interface{}),
		}

		// 解析属性
		props, err := row.GetValueByColName("Properties")
		if err == nil {
			if propsMap, err := props.AsMap(); err == nil {
				for k, v := range propsMap {
					vertex.Props[k] = valueWrapperToInterface(v)
				}
			}
		}

		return vertex, nil
	}

	return nil, fmt.Errorf("vertex not found: %s", vid)
}

// Update 更新顶点属性
// UPDATE VERTEX ON tag_name vid SET prop1 = value1, prop2 = value2
func (s *VertexService) Update(ctx context.Context, vertex Vertex) error {
	if vertex.VID == "" || vertex.Tag == "" {
		return fmt.Errorf("VID and Tag are required")
	}

	var updates []string
	for key, value := range vertex.Props {
		updates = append(updates, fmt.Sprintf("%s = %s", key, formatValue(value)))
	}

	if len(updates) == 0 {
		return fmt.Errorf("no properties to update")
	}

	stmt := fmt.Sprintf(
		"UPDATE VERTEX ON %s \"%s\" SET %s",
		vertex.Tag,
		vertex.VID,
		strings.Join(updates, ", "),
	)

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("update vertex failed",
			zap.String("tag", vertex.Tag),
			zap.String("vid", vertex.VID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("vertex updated",
		zap.String("tag", vertex.Tag),
		zap.String("vid", vertex.VID),
	)
	return nil
}

// Delete 删除顶点
// DELETE VERTEX vid
func (s *VertexService) Delete(ctx context.Context, vid string, withEdge bool) error {
	var stmt string
	if withEdge {
		stmt = fmt.Sprintf("DELETE VERTEX \"%s\" WITH EDGE", vid)
	} else {
		stmt = fmt.Sprintf("DELETE VERTEX \"%s\"", vid)
	}

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("delete vertex failed",
			zap.String("vid", vid),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("vertex deleted",
		zap.String("vid", vid),
	)
	return nil
}

// Query 自定义查询顶点
func (s *VertexService) Query(ctx context.Context, query string) (*nebula_go.ResultSet, error) {
	return s.client.Execute(ctx, query)
}

// formatValue 格式化属性值
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", escapeString(v))
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

// escapeString 转义字符串中的特殊字符
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
