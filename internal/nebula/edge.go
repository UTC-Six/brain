package nebula

import (
	"context"
	"fmt"
	"strings"

	nebula_go "github.com/vesoft-inc/nebula-go/v3"
	"go.uber.org/zap"
)

// Edge 边数据
type Edge struct {
	SrcID string                 // 起始顶点ID
	DstID string                 // 目标顶点ID
	Type  string                 // 边类型名称
	Rank  int64                  // 边排名（可选，默认为0）
	Props map[string]interface{} // 属性
}

// EdgeService 边操作服务
type EdgeService struct {
	client *Client
	logger *zap.Logger
}

// NewEdgeService 创建边服务
func NewEdgeService(client *Client, logger *zap.Logger) *EdgeService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EdgeService{
		client: client,
		logger: logger,
	}
}

// Create 创建边
// INSERT EDGE edge_type (prop1, prop2, ...) VALUES src_id -> dst_id:(value1, value2, ...)
func (s *EdgeService) Create(ctx context.Context, edge Edge) error {
	if edge.SrcID == "" || edge.DstID == "" || edge.Type == "" {
		return fmt.Errorf("SrcID, DstID and Type are required")
	}

	var props []string
	var values []string
	for key, value := range edge.Props {
		props = append(props, key)
		values = append(values, formatValue(value))
	}

	var stmt string
	if edge.Rank != 0 {
		stmt = fmt.Sprintf(
			"INSERT EDGE %s (%s) VALUES \"%s\" -> \"%s\"@%d:(%s)",
			edge.Type,
			strings.Join(props, ", "),
			edge.SrcID,
			edge.DstID,
			edge.Rank,
			strings.Join(values, ", "),
		)
	} else {
		stmt = fmt.Sprintf(
			"INSERT EDGE %s (%s) VALUES \"%s\" -> \"%s\":(%s)",
			edge.Type,
			strings.Join(props, ", "),
			edge.SrcID,
			edge.DstID,
			strings.Join(values, ", "),
		)
	}

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("create edge failed",
			zap.String("type", edge.Type),
			zap.String("src", edge.SrcID),
			zap.String("dst", edge.DstID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("edge created",
		zap.String("type", edge.Type),
		zap.String("src", edge.SrcID),
		zap.String("dst", edge.DstID),
	)
	return nil
}

// BatchCreate 批量创建边
func (s *EdgeService) BatchCreate(ctx context.Context, edges []Edge) error {
	if len(edges) == 0 {
		return nil
	}

	// 按类型分组
	typeGroups := make(map[string][]Edge)
	for _, e := range edges {
		typeGroups[e.Type] = append(typeGroups[e.Type], e)
	}

	// 为每种类型批量插入
	for edgeType, es := range typeGroups {
		if err := s.batchCreateByType(ctx, edgeType, es); err != nil {
			return err
		}
	}

	return nil
}

// batchCreateByType 按类型批量创建
func (s *EdgeService) batchCreateByType(ctx context.Context, edgeType string, edges []Edge) error {
	if len(edges) == 0 {
		return nil
	}

	// 获取第一条边的属性键作为模板
	firstEdge := edges[0]
	var props []string
	for key := range firstEdge.Props {
		props = append(props, key)
	}

	// 构建批量插入语句
	var valueParts []string
	for _, e := range edges {
		var values []string
		for _, prop := range props {
			value := firstEdge.Props[prop]
			if val, ok := e.Props[prop]; ok {
				value = val
			}
			values = append(values, formatValue(value))
		}

		if e.Rank != 0 {
			valueParts = append(valueParts, fmt.Sprintf(
				"\"%s\" -> \"%s\"@%d:(%s)",
				e.SrcID, e.DstID, e.Rank, strings.Join(values, ", "),
			))
		} else {
			valueParts = append(valueParts, fmt.Sprintf(
				"\"%s\" -> \"%s\":(%s)",
				e.SrcID, e.DstID, strings.Join(values, ", "),
			))
		}
	}

	stmt := fmt.Sprintf(
		"INSERT EDGE %s (%s) VALUES %s",
		edgeType,
		strings.Join(props, ", "),
		strings.Join(valueParts, ", "),
	)

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("batch create edges failed",
			zap.String("type", edgeType),
			zap.Int("count", len(edges)),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("edges batch created",
		zap.String("type", edgeType),
		zap.Int("count", len(edges)),
	)
	return nil
}

// Get 查询边
// FETCH PROP ON edge_type src_id -> dst_id
func (s *EdgeService) Get(ctx context.Context, edgeType string, srcID string, dstID string, rank int64) (*Edge, error) {
	var stmt string
	if rank != 0 {
		stmt = fmt.Sprintf("FETCH PROP ON %s \"%s\" -> \"%s\"@%d", edgeType, srcID, dstID, rank)
	} else {
		stmt = fmt.Sprintf("FETCH PROP ON %s \"%s\" -> \"%s\"", edgeType, srcID, dstID)
	}

	resultSet, err := s.client.Execute(ctx, stmt)
	if err != nil {
		return nil, err
	}

	if !resultSet.IsEmpty() {
		row, err := resultSet.GetRowValuesByIndex(0)
		if err != nil {
			return nil, err
		}

		edge := &Edge{
			SrcID: srcID,
			DstID: dstID,
			Type:  edgeType,
			Rank:  rank,
			Props: make(map[string]interface{}),
		}

		// 解析属性
		props, err := row.GetValueByColName("Properties")
		if err == nil {
			if propsMap, err := props.AsMap(); err == nil {
				for k, v := range propsMap {
					edge.Props[k] = valueWrapperToInterface(v)
				}
			}
		}

		return edge, nil
	}

	return nil, fmt.Errorf("edge not found: %s -> %s", srcID, dstID)
}

// Update 更新边属性
// UPDATE EDGE ON edge_type src_id -> dst_id SET prop1 = value1, prop2 = value2
func (s *EdgeService) Update(ctx context.Context, edge Edge) error {
	if edge.SrcID == "" || edge.DstID == "" || edge.Type == "" {
		return fmt.Errorf("SrcID, DstID and Type are required")
	}

	var updates []string
	for key, value := range edge.Props {
		updates = append(updates, fmt.Sprintf("%s = %s", key, formatValue(value)))
	}

	if len(updates) == 0 {
		return fmt.Errorf("no properties to update")
	}

	var stmt string
	if edge.Rank != 0 {
		stmt = fmt.Sprintf(
			"UPDATE EDGE ON %s \"%s\" -> \"%s\"@%d SET %s",
			edge.Type,
			edge.SrcID,
			edge.DstID,
			edge.Rank,
			strings.Join(updates, ", "),
		)
	} else {
		stmt = fmt.Sprintf(
			"UPDATE EDGE ON %s \"%s\" -> \"%s\" SET %s",
			edge.Type,
			edge.SrcID,
			edge.DstID,
			strings.Join(updates, ", "),
		)
	}

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("update edge failed",
			zap.String("type", edge.Type),
			zap.String("src", edge.SrcID),
			zap.String("dst", edge.DstID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("edge updated",
		zap.String("type", edge.Type),
		zap.String("src", edge.SrcID),
		zap.String("dst", edge.DstID),
	)
	return nil
}

// Delete 删除边
// DELETE EDGE edge_type src_id -> dst_id
func (s *EdgeService) Delete(ctx context.Context, edgeType string, srcID string, dstID string, rank int64) error {
	var stmt string
	if rank != 0 {
		stmt = fmt.Sprintf("DELETE EDGE %s \"%s\" -> \"%s\"@%d", edgeType, srcID, dstID, rank)
	} else {
		stmt = fmt.Sprintf("DELETE EDGE %s \"%s\" -> \"%s\"", edgeType, srcID, dstID)
	}

	_, err := s.client.Execute(ctx, stmt)
	if err != nil {
		s.logger.Error("delete edge failed",
			zap.String("type", edgeType),
			zap.String("src", srcID),
			zap.String("dst", dstID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("edge deleted",
		zap.String("type", edgeType),
		zap.String("src", srcID),
		zap.String("dst", dstID),
	)
	return nil
}

// Query 自定义查询边
func (s *EdgeService) Query(ctx context.Context, query string) (*nebula_go.ResultSet, error) {
	return s.client.Execute(ctx, query)
}
