package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

// VectorService 向量操作服务
type VectorService struct {
	client *Client
	logger *zap.Logger
}

// NewVectorService 创建向量服务
func NewVectorService(client *Client, logger *zap.Logger) *VectorService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &VectorService{
		client: client,
		logger: logger,
	}
}

// Insert 插入向量数据
func (s *VectorService) Insert(ctx context.Context, collectionName string, data []map[string]interface{}) (entity.Column, error) {
	milvusClient := s.client.GetClient()

	// 转换为 Milvus 的列格式
	columns := make([]entity.Column, 0)
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	// 获取第一个数据项的键作为字段名
	firstItem := data[0]
	fieldMap := make(map[string][]interface{})

	for key := range firstItem {
		fieldMap[key] = make([]interface{}, 0, len(data))
	}

	// 填充数据
	for _, item := range data {
		for key, value := range item {
			if _, ok := fieldMap[key]; ok {
				fieldMap[key] = append(fieldMap[key], value)
			}
		}
	}

	// 构建列
	for fieldName, values := range fieldMap {
		// 判断第一个值的类型
		if len(values) == 0 {
			continue
		}

		switch v := values[0].(type) {
		case []float32:
			// 向量字段
			vectors := make([][]float32, 0, len(values))
			for _, val := range values {
				if vec, ok := val.([]float32); ok {
					vectors = append(vectors, vec)
				}
			}
			if len(vectors) == 0 || len(vectors[0]) == 0 {
				return nil, fmt.Errorf("float vector field %s is empty", fieldName)
			}
			column := entity.NewColumnFloatVector(fieldName, len(vectors[0]), vectors)
			columns = append(columns, column)
		case int64:
			// Int64 字段
			intValues := make([]int64, 0, len(values))
			for _, val := range values {
				if intVal, ok := val.(int64); ok {
					intValues = append(intValues, intVal)
				}
			}
			column := entity.NewColumnInt64(fieldName, intValues)
			columns = append(columns, column)
		case string:
			// String 字段
			strValues := make([]string, 0, len(values))
			for _, val := range values {
				if strVal, ok := val.(string); ok {
					strValues = append(strValues, strVal)
				}
			}
			column := entity.NewColumnVarChar(fieldName, strValues)
			columns = append(columns, column)
		case float64:
			// Double 字段
			floatValues := make([]float64, 0, len(values))
			for _, val := range values {
				if floatVal, ok := val.(float64); ok {
					floatValues = append(floatValues, floatVal)
				}
			}
			column := entity.NewColumnDouble(fieldName, floatValues)
			columns = append(columns, column)
		case bool:
			// Bool 字段
			boolValues := make([]bool, 0, len(values))
			for _, val := range values {
				if boolVal, ok := val.(bool); ok {
					boolValues = append(boolValues, boolVal)
				}
			}
			column := entity.NewColumnBool(fieldName, boolValues)
			columns = append(columns, column)
		default:
			return nil, fmt.Errorf("unsupported field type for field %s: %T", fieldName, v)
		}
	}

	// 插入数据
	ids, err := milvusClient.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		s.logger.Error("insert vectors failed",
			zap.String("collection", collectionName),
			zap.Int("count", len(data)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("insert vectors failed: %w", err)
	}

	s.logger.Info("vectors inserted",
		zap.String("collection", collectionName),
		zap.Int("count", len(data)),
		zap.Int("ids_count", ids.Len()),
	)

	return ids, nil
}

// Search 向量相似度搜索
func (s *VectorService) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
	milvusClient := s.client.GetClient()

	// 构建搜索参数
	searchParams := newGenericSearchParam(req.SearchParams)

	// 转换查询向量
	vectors := make([]entity.Vector, 0, len(req.Vectors))
	for _, vec := range req.Vectors {
		vectors = append(vectors, entity.FloatVector(vec))
	}

	// 执行搜索
	searchResults, err := milvusClient.Search(
		ctx,
		req.CollectionName,
		req.Partitions,
		req.Expr,
		req.OutputFields,
		vectors,
		req.VectorField,
		req.MetricType,
		req.TopK,
		searchParams,
	)
	if err != nil {
		s.logger.Error("search vectors failed",
			zap.String("collection", req.CollectionName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search vectors failed: %w", err)
	}

	results := make([]*SearchResult, 0, len(searchResults))
	for _, sr := range searchResults {
		results = append(results, &SearchResult{
			ResultCount: sr.ResultCount,
			IDs:         sr.IDs,
			Fields:      sr.Fields,
			Scores:      sr.Scores,
			Err:         sr.Err,
		})
	}

	s.logger.Info("vectors searched",
		zap.String("collection", req.CollectionName),
		zap.Int("results_count", len(results)),
	)

	return results, nil
}

// Delete 删除向量数据
func (s *VectorService) Delete(ctx context.Context, collectionName string, expr string) error {
	milvusClient := s.client.GetClient()

	err := milvusClient.Delete(ctx, collectionName, "", expr)
	if err != nil {
		s.logger.Error("delete vectors failed",
			zap.String("collection", collectionName),
			zap.String("expr", expr),
			zap.Error(err),
		)
		return fmt.Errorf("delete vectors failed: %w", err)
	}

	s.logger.Info("vectors deleted",
		zap.String("collection", collectionName),
		zap.String("expr", expr),
	)
	return nil
}

// Query 查询向量数据（非向量搜索）
func (s *VectorService) Query(ctx context.Context, req *QueryRequest) ([]entity.Column, error) {
	milvusClient := s.client.GetClient()

	resultSet, err := milvusClient.Query(
		ctx,
		req.CollectionName,
		req.Partitions,
		req.Expr,
		req.OutputFields,
	)
	if err != nil {
		s.logger.Error("query vectors failed",
			zap.String("collection", req.CollectionName),
			zap.String("expr", req.Expr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("query vectors failed: %w", err)
	}

	s.logger.Info("vectors queried",
		zap.String("collection", req.CollectionName),
		zap.Int("field_count", len(resultSet)),
	)

	return resultSet, nil
}

// SearchRequest 搜索请求
type SearchRequest struct {
	CollectionName string
	Partitions     []string
	Expr           string
	OutputFields   []string
	VectorField    string
	Vectors        [][]float32
	MetricType     entity.MetricType
	TopK           int
	SearchParams   map[string]interface{}
}

// SearchResult 搜索结果
type SearchResult struct {
	ResultCount int
	IDs         entity.Column
	Fields      []entity.Column
	Scores      []float32
	Err         error
}

// QueryRequest 查询请求
type QueryRequest struct {
	CollectionName string
	Partitions     []string
	Expr           string
	OutputFields   []string
}

// genericSearchParam implements entity.SearchParam with dynamic params map.
type genericSearchParam struct {
	params map[string]interface{}
}

func newGenericSearchParam(values map[string]interface{}) entity.SearchParam {
	p := make(map[string]interface{})
	for k, v := range values {
		p[k] = v
	}
	return &genericSearchParam{params: p}
}

func (g *genericSearchParam) Params() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range g.params {
		result[k] = v
	}
	return result
}

func (g *genericSearchParam) AddRadius(radius float64) {
	g.params["radius"] = radius
}

func (g *genericSearchParam) AddRangeFilter(rangeFilter float64) {
	g.params["range_filter"] = rangeFilter
}
