# Brain - Nebula Graph & Milvus 数据库封装

一个专为 AI 教育场景设计的 Golang 数据库封装库，提供 Nebula Graph（图数据库）和 Milvus（向量数据库）的统一访问接口。

## 📋 目录

- [特性](#特性)
- [架构设计](#架构设计)
- [快速开始](#快速开始)
- [最佳实践](#最佳实践)
- [API 文档](#api-文档)
- [使用示例](#使用示例)
- [配置说明](#配置说明)

## ✨ 特性

### Nebula Graph 封装
- ✅ 连接池管理（支持最大/最小连接数配置）
- ✅ 顶点（Vertex）CRUD 操作
- ✅ 边（Edge）CRUD 操作
- ✅ 批量操作支持
- ✅ 自动重试机制
- ✅ 查询结果解析封装
- ✅ 连接健康检查

### Milvus 封装
- ✅ 连接管理（支持 TLS 和认证）
- ✅ Collection 生命周期管理
- ✅ 向量插入、查询、删除
- ✅ 相似度搜索（支持多种距离度量）
- ✅ 索引管理（HNSW、IVF_FLAT、FLAT）
- ✅ 动态字段支持（存储元数据）
- ✅ 集合统计信息查询

### 业务场景支持
- 🎓 学生-课程-知识点关系图谱
- 📊 学习记录分析和统计
- 🔍 学习内容向量召回
- 📈 学习能力评估
- 🔗 知识点关联关系查询

## 🏗️ 架构设计

### 设计原则

1. **分层架构**
   - `internal/`: 核心封装实现（不对外暴露）
   - `pkg/`: 公共包（配置、统一接口）
   - `examples/`: 使用示例和业务场景封装

2. **服务化设计**
   - 每个数据库提供独立的 Service 层
   - 支持依赖注入和测试
   - 统一的错误处理和日志记录

3. **最佳实践**
   - 连接池管理（避免连接泄漏）
   - 上下文传递（支持超时和取消）
   - 结构化日志（zap）
   - 配置外部化（viper）

### 目录结构

```
brain/
├── internal/
│   ├── nebula/          # Nebula Graph 封装
│   │   ├── client.go    # 客户端和连接管理
│   │   ├── vertex.go    # 顶点操作
│   │   └── edge.go      # 边操作
│   └── milvus/          # Milvus 封装
│       ├── client.go    # 客户端和连接管理
│       ├── collection.go # Collection 管理
│       ├── vector.go    # 向量操作
│       └── index.go     # 索引管理
├── pkg/
│   ├── config/          # 配置管理
│   └── brain/           # 统一访问接口
├── examples/
│   ├── education/       # AI 教育场景示例
│   └── main.go          # 完整使用示例
├── config/
│   └── config.yaml      # 配置文件示例
└── README.md
```

## 🚀 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置数据库

编辑 `config/config.yaml`，配置 Nebula Graph 和 Milvus 的连接信息。

### 3. 初始化 Nebula Graph 图空间

```bash
# 连接到 Nebula Graph
nebula-console -addr 127.0.0.1 -port 9669 -u root -p password

# 创建图空间
CREATE SPACE IF NOT EXISTS brain(partition_num=10, replica_factor=1);
USE brain;
```

### 4. 运行示例

```bash
go run examples/main.go
```

## 💡 最佳实践

### 1. Nebula Graph 使用场景

**适用场景：**
- 学生-课程-知识点关系图谱
- 知识点先修关系（prerequisite）
- 学习路径推荐
- 社交网络分析

**最佳实践：**
- 使用 Tag 表示实体类型（student, course, knowledge_point）
- 使用 Edge 表示关系（studies, learns, contains, prerequisite）
- 为常用查询创建索引
- 批量操作时按 Tag/Edge 类型分组

**示例：**
```go
// 创建学生顶点
vertex := nebula.Vertex{
    VID: "student_001",
    Tag: "student",
    Props: map[string]interface{}{
        "id":   "student_001",
        "name": "张三",
        "age":  15,
    },
}
vertexService.Create(ctx, vertex)

// 创建学习关系边
edge := nebula.Edge{
    SrcID: "student_001",
    DstID: "kp_001",
    Type:  "learns",
    Props: map[string]interface{}{
        "score":        85.5,
        "mastery_level": 4,
    },
}
edgeService.Create(ctx, edge)
```

### 2. Milvus 使用场景

**适用场景：**
- 学习内容向量化存储（题目、答案、材料）
- 相似内容召回（推荐系统）
- 语义搜索
- 知识图谱向量化

**最佳实践：**
- 使用 HNSW 索引（适合高维向量，查询速度快）
- 向量维度与 embedding 模型保持一致
- 使用动态字段存储元数据
- 定期优化索引参数（M, efConstruction）

**示例：**
```go
// 创建向量集合
schema := &entity.Schema{
    CollectionName: "learning_content",
    Fields: []*entity.Field{
        {Name: "id", DataType: entity.FieldTypeVarChar, PrimaryKey: true},
        {Name: "vector", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "128"}},
        {Name: "knowledge_id", DataType: entity.FieldTypeVarChar},
    },
    EnableDynamicField: true,
}
collectionService.CreateCollection(ctx, schema)

// 插入向量
data := []map[string]interface{}{
    {
        "id":          "content_001",
        "vector":      []float32{...}, // 通过 embedding 模型生成
        "knowledge_id": "kp_001",
    },
}
vectorService.Insert(ctx, "learning_content", data)
```

### 3. 组合使用场景

**学习能力分析：**
1. 从 Nebula Graph 查询学生的学习记录（图查询）
2. 从 Milvus 召回相关的学习内容（向量搜索）
3. 结合两者结果进行综合分析

**内容推荐：**
1. 从 Nebula Graph 获取学生的薄弱知识点
2. 在 Milvus 中搜索这些知识点的相似内容
3. 返回推荐内容列表

## 📚 API 文档

### Nebula Graph

#### Client
- `NewClient(config, logger)` - 创建客户端
- `Execute(ctx, stmt)` - 执行 nGQL 查询
- `Ping(ctx)` - 健康检查
- `Close()` - 关闭连接

#### VertexService
- `Create(ctx, vertex)` - 创建顶点
- `BatchCreate(ctx, vertices)` - 批量创建
- `Get(ctx, tag, vid)` - 查询顶点
- `Update(ctx, vertex)` - 更新顶点
- `Delete(ctx, vid, withEdge)` - 删除顶点

#### EdgeService
- `Create(ctx, edge)` - 创建边
- `BatchCreate(ctx, edges)` - 批量创建
- `Get(ctx, edgeType, srcID, dstID, rank)` - 查询边
- `Update(ctx, edge)` - 更新边
- `Delete(ctx, edgeType, srcID, dstID, rank)` - 删除边

### Milvus

#### Client
- `NewClient(config, logger)` - 创建客户端
- `Ping(ctx)` - 健康检查
- `Close()` - 关闭连接

#### CollectionService
- `CreateCollection(ctx, schema)` - 创建集合
- `DropCollection(ctx, name)` - 删除集合
- `HasCollection(ctx, name)` - 检查集合是否存在
- `LoadCollection(ctx, name)` - 加载集合到内存
- `GetCollectionStats(ctx, name)` - 获取统计信息

#### VectorService
- `Insert(ctx, collectionName, data)` - 插入向量
- `Search(ctx, req)` - 向量相似度搜索
- `Query(ctx, req)` - 标量查询
- `Delete(ctx, collectionName, expr)` - 删除向量

#### IndexService
- `CreateIndex(ctx, collectionName, fieldName, index)` - 创建索引
- `DropIndex(ctx, collectionName, fieldName)` - 删除索引
- `DescribeIndex(ctx, collectionName, fieldName)` - 描述索引

## 📖 使用示例

### 完整示例

参考 `examples/main.go` 查看完整的使用示例，包括：
- 初始化 Schema
- 创建实体和关系
- 查询和分析
- 向量存储和搜索

### AI 教育场景示例

参考 `examples/education/education.go` 查看业务场景封装，包括：
- 学生、课程、知识点管理
- 学习记录分析
- 学习能力评估
- 内容向量召回

## ⚙️ 配置说明

### Nebula Graph 配置

```yaml
nebula:
  addresses:        # GraphD 地址列表
    - "127.0.0.1:9669"
  username: "root"  # 用户名
  password: "password"  # 密码
  space: "brain"    # 图空间名称
  timeout: 10s      # 连接超时
  max_conn: 10      # 最大连接数
  min_conn: 2       # 最小连接数
```

### Milvus 配置

```yaml
milvus:
  host: "127.0.0.1"  # 服务器地址
  port: 19530        # 端口号
  username: ""       # 用户名（可选）
  password: ""       # 密码（可选）
  database: "default"  # 数据库名称
  connect_timeout: 10s  # 连接超时
  enable_tls: false     # 是否启用 TLS
```

## 🔧 开发建议

### 1. 连接管理
- 使用连接池避免频繁创建连接
- 在生产环境中设置合理的连接数
- 实现健康检查和自动重连

### 2. 性能优化
- 批量操作时使用 BatchCreate
- 为常用查询字段创建索引
- 使用合适的向量索引类型（HNSW 推荐）

### 3. 错误处理
- 所有操作都返回 error，需要检查
- 使用结构化日志记录错误
- 实现重试机制处理临时错误

### 4. 测试
- 使用测试数据库进行单元测试
- 模拟网络错误和超时场景
- 测试并发安全性

## 📝 最佳实践理由

### 为什么选择 Nebula Graph？
1. **原生图数据库**：专为图数据设计，查询性能优异
2. **nGQL 查询语言**：类似 SQL，学习成本低
3. **分布式架构**：支持水平扩展
4. **社区活跃**：Vesoft 公司维护，文档完善

### 为什么选择 Milvus？
1. **向量数据库**：专为向量搜索优化
2. **高性能**：支持大规模向量检索
3. **多种索引**：HNSW、IVF 等，适应不同场景
4. **云原生**：支持 Kubernetes 部署

### 为什么这样封装？
1. **统一接口**：简化业务代码，降低学习成本
2. **类型安全**：Go 强类型，减少运行时错误
3. **可测试性**：接口化设计，易于 mock
4. **可扩展性**：模块化设计，易于扩展新功能

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

