# KingClub AI 教育平台 - Nebula Graph & Milvus 数据架构设计

## 一、设计目标

通过 Nebula Graph 存储学生关系数据和事件流，通过 Milvus 存储向量化内容（会话、文章、问答等），实现：

1. **快速检索**：通过向量相似度搜索召回相关内容
2. **关系分析**：通过图查询分析学生关系网络和学习路径
3. **画像构建**：结合图数据和向量数据，为大模型提供全面的学生画像材料

## 二、Nebula Graph Schema 设计

### 2.1 顶点（Tag）定义

#### Student（学生）

```ngql
CREATE TAG IF NOT EXISTS student(
    id string,                    -- 学生ID
    name string,                   -- 姓名
    age int,                       -- 年龄
    grade string,                  -- 年级
    gender string,                 -- 性别
    avatar_url string,             -- 头像URL
    phone string,                  -- 手机号
    email string,                  -- 邮箱
    school string,                 -- 学校
    class_name string,             -- 班级
    enrollment_date timestamp,     -- 入学日期
    create_at timestamp,           -- 创建时间
    update_at timestamp            -- 更新时间
);
```

#### Course（课程）

```ngql
CREATE TAG IF NOT EXISTS course(
    id string,                     -- 课程ID
    name string,                    -- 课程名称
    subject string,                 -- 学科（语文、数学、英语等）
    level string,                   -- 难度等级
    description string,             -- 课程描述
    duration int,                   -- 课程时长（分钟）
    create_at timestamp
);
```

#### Lesson（课时）

```ngql
CREATE TAG IF NOT EXISTS lesson(
    id string,                      -- 课时ID
    course_id string,               -- 所属课程ID
    title string,                    -- 课时标题
    lesson_type string,             -- 课时类型（写作课、阅读课等）
    start_time timestamp,           -- 开始时间
    end_time timestamp,             -- 结束时间
    create_at timestamp
);
```

#### KnowledgePoint（知识点）

```ngql
CREATE TAG IF NOT EXISTS knowledge_point(
    id string,                      -- 知识点ID
    name string,                     -- 知识点名称
    subject string,                  -- 所属学科
    difficulty int,                  -- 难度（1-5）
    description string,              -- 描述
    create_at timestamp
);
```

#### Article（文章）

```ngql
CREATE TAG IF NOT EXISTS article(
    id string,                       -- 文章ID
    student_id string,               -- 学生ID
    lesson_id string,                -- 课时ID
    title string,                     -- 标题
    content text,                    -- 内容
    word_count int,                  -- 字数
    score double,                    -- 评分
    feedback text,                   -- 反馈
    create_at timestamp
);
```

#### Question（题目）

```ngql
CREATE TAG IF NOT EXISTS question(
    id string,                       -- 题目ID
    lesson_id string,                -- 课时ID
    question_type string,            -- 题型（选择题、填空题等）
    content text,                    -- 题目内容
    correct_answer string,           -- 正确答案
    difficulty int,                  -- 难度
    create_at timestamp
);
```

#### Conversation（会话）

```ngql
CREATE TAG IF NOT EXISTS conversation(
    id string,                       -- 会话ID
    student_id string,               -- 学生ID
    lesson_id string,                -- 课时ID（可为空，首页会话）
    conversation_type string,        -- 会话类型（首页、课中、课后）
    summary text,                    -- 会话摘要
    message_count int,               -- 消息数量
    start_time timestamp,            -- 开始时间
    end_time timestamp,              -- 结束时间
    create_at timestamp
);
```

#### FamilyMember（家庭成员）

```ngql
CREATE TAG IF NOT EXISTS family_member(
    id string,                       -- 成员ID
    name string,                     -- 姓名
    relationship string,              -- 关系（父亲、母亲、祖父等）
    education_level string,          -- 教育水平
    occupation string,               -- 职业
    create_at timestamp
);
```

#### Teacher（老师）

```ngql
CREATE TAG IF NOT EXISTS teacher(
    id string,                       -- 老师ID
    name string,                     -- 姓名
    subject string,                  -- 教学科目
    experience_years int,           -- 教学经验（年）
    create_at timestamp
);
```

#### Hobby（兴趣爱好）

```ngql
CREATE TAG IF NOT EXISTS hobby(
    id string,                       -- 兴趣ID
    name string,                     -- 兴趣名称
    category string,                 -- 类别（运动、艺术、科技等）
    create_at timestamp
);
```

#### MBTITest（MBTI测试）

```ngql
CREATE TAG IF NOT EXISTS mbti_test(
    id string,                       -- 测试ID
    student_id string,               -- 学生ID
    mbti_type string,                -- MBTI类型（INTJ、ENFP等）
    e_score int,                     -- 外向性得分
    i_score int,                     -- 内向性得分
    s_score int,                     -- 感觉得分
    n_score int,                     -- 直觉得分
    t_score int,                     -- 思考得分
    f_score int,                     -- 情感得分
    j_score int,                     -- 判断得分
    p_score int,                     -- 感知得分
    test_date timestamp,             -- 测试日期
    create_at timestamp
);
```

### 2.2 边（Edge）定义

#### 学习相关边

```ngql
-- 学生参加课程
CREATE EDGE IF NOT EXISTS attends(
    student_id string,
    course_id string,
    enroll_time timestamp,           -- 报名时间
    progress double,                  -- 学习进度（0-1）
    status string                     -- 状态（进行中、已完成、已退课）
);

-- 学生学习知识点
CREATE EDGE IF NOT EXISTS studies(
    student_id string,
    knowledge_id string,
    mastery_level int,                -- 掌握程度（1-5）
    study_count int,                  -- 学习次数
    last_study_time timestamp,        -- 最后学习时间
    total_study_duration int          -- 总学习时长（分钟）
);

-- 课时包含知识点
CREATE EDGE IF NOT EXISTS contains(
    lesson_id string,
    knowledge_id string,
    order_index int                   -- 顺序
);

-- 知识点前置关系
CREATE EDGE IF NOT EXISTS prerequisite(
    knowledge_id1 string,             -- 前置知识点
    knowledge_id2 string              -- 后续知识点
);
```

#### 内容创作边

```ngql
-- 学生写文章
CREATE EDGE IF NOT EXISTS writes(
    student_id string,
    article_id string,
    lesson_id string,
    write_time timestamp,
    word_count int,
    score double
);

-- 学生答题
CREATE EDGE IF NOT EXISTS answers(
    student_id string,
    question_id string,
    lesson_id string,
    answer_content string,            -- 答案内容
    is_correct bool,                 -- 是否正确
    answer_time timestamp,
    time_spent int                   -- 答题耗时（秒）
);
```

#### 会话边

```ngql
-- 学生会话
CREATE EDGE IF NOT EXISTS chats(
    student_id string,
    conversation_id string,
    lesson_id string,
    message_count int,
    avg_response_time double,         -- 平均响应时间（秒）
    engagement_score double           -- 参与度得分（0-1）
);
```

#### 关系边

```ngql
-- 学生属于家庭
CREATE EDGE IF NOT EXISTS belongs_to(
    student_id string,
    family_member_id string,
    relationship string,
    closeness int                     -- 亲密度（1-5）
);

-- 学生与同学关系
CREATE EDGE IF NOT EXISTS friends_with(
    student_id1 string,
    student_id2 string,
    friendship_level int,            -- 友谊程度（1-5）
    since timestamp                   -- 成为朋友时间
);

-- 学生喜欢兴趣
CREATE EDGE IF NOT EXISTS likes(
    student_id string,
    hobby_id string,
    interest_level int,               -- 兴趣程度（1-5）
    since timestamp
);

-- 学生有MBTI类型
CREATE EDGE IF NOT EXISTS has_mbti(
    student_id string,
    mbti_test_id string,
    mbti_type string,
    test_date timestamp
);
```

#### 学习习惯边

```ngql
-- 学生预习
CREATE EDGE IF NOT EXISTS previews(
    student_id string,
    lesson_id string,
    preview_time timestamp,
    preview_duration int,             -- 预习时长（分钟）
    completion_rate double            -- 完成率（0-1）
);

-- 学生复习
CREATE EDGE IF NOT EXISTS reviews(
    student_id string,
    lesson_id string,
    review_time timestamp,
    review_duration int,              -- 复习时长（分钟）
    review_count int                  -- 复习次数
);

-- 学生完成作业
CREATE EDGE IF NOT EXISTS completes_homework(
    student_id string,
    lesson_id string,
    homework_id string,
    submit_time timestamp,
    completion_time int,             -- 完成耗时（分钟）
    score double,                     -- 作业得分
    is_on_time bool                  -- 是否按时提交
);

-- 学生整理错题
CREATE EDGE IF NOT EXISTS collects_mistake(
    student_id string,
    question_id string,
    collect_time timestamp,
    review_count int,                -- 复习次数
    is_mastered bool                  -- 是否已掌握
);
```

#### 课堂行为边

```ngql
-- 学生课堂听讲状态
CREATE EDGE IF NOT EXISTS listens(
    student_id string,
    lesson_id string,
    attention_score double,           -- 注意力得分（0-1）
    interaction_count int,            -- 互动次数
    focus_duration int                -- 专注时长（分钟）
);

-- 学生课堂互动
CREATE EDGE IF NOT EXISTS interacts(
    student_id string,
    lesson_id string,
    interaction_type string,          -- 互动类型（提问、回答、讨论等）
    interaction_time timestamp,
    quality_score double              -- 互动质量得分（0-1）
);
```

## 三、Milvus 集合设计

### 3.1 会话向量集合（conversation_vectors）

**用途**：存储学生与大模型的会话内容向量，用于语义检索和对话分析

**Schema**：

```go
Fields: [
{Name: "id", DataType: VarChar, MaxLength: 100, IsPrimaryKey: true},
{Name: "vector", DataType: FloatVector, Dim: 768}, // 使用中文embedding模型
{Name: "conversation_id", DataType: VarChar, MaxLength: 100},
{Name: "student_id", DataType: VarChar, MaxLength: 100},
{Name: "lesson_id", DataType: VarChar, MaxLength: 100},
{Name: "message_type", DataType: VarChar, MaxLength: 20}, // user/assistant
{Name: "content", DataType: VarChar, MaxLength: 2000},    // 消息内容
{Name: "timestamp", DataType: Int64},
{Name: "conversation_type", DataType: VarChar, MaxLength: 20}, // 首页/课中/课后
]
EnableDynamicField: true // 支持存储额外metadata
```

**索引**：HNSW (L2, M=16, efConstruction=200)

### 3.2 文章向量集合（article_vectors）

**用途**：存储学生写作课文章向量，用于文章质量分析和相似文章检索

**Schema**：

```go
Fields: [
{Name: "id", DataType: VarChar, MaxLength: 100, IsPrimaryKey: true},
{Name: "vector", DataType: FloatVector, Dim: 768},
{Name: "article_id", DataType: VarChar, MaxLength: 100},
{Name: "student_id", DataType: VarChar, MaxLength: 100},
{Name: "lesson_id", DataType: VarChar, MaxLength: 100},
{Name: "title", DataType: VarChar, MaxLength: 200},
{Name: "content", DataType: VarChar, MaxLength: 10000},
{Name: "word_count", DataType: Int32},
{Name: "score", DataType: Double},
{Name: "timestamp", DataType: Int64},
]
EnableDynamicField: true
```

**索引**：HNSW (L2, M=16, efConstruction=200)

### 3.3 问答向量集合（question_answer_vectors）

**用途**：存储阅读课的问答记录，用于理解学生答题思路和知识掌握情况

**Schema**：

```go
Fields: [
{Name: "id", DataType: VarChar, MaxLength: 100, IsPrimaryKey: true},
{Name: "vector", DataType: FloatVector, Dim: 768},
{Name: "question_id", DataType: VarChar, MaxLength: 100},
{Name: "student_id", DataType: VarChar, MaxLength: 100},
{Name: "lesson_id", DataType: VarChar, MaxLength: 100},
{Name: "question_content", DataType: VarChar, MaxLength: 1000},
{Name: "answer_content", DataType: VarChar, MaxLength: 2000},
{Name: "is_correct", DataType: Bool},
{Name: "time_spent", DataType: Int32}, // 秒
{Name: "timestamp", DataType: Int64},
]
EnableDynamicField: true
```

**索引**：HNSW (L2, M=16, efConstruction=200)

### 3.4 学习行为向量集合（learning_behavior_vectors）

**用途**：存储学习行为序列的向量表示，用于行为模式识别和习惯分析

**Schema**：

```go
Fields: [
{Name: "id", DataType: VarChar, MaxLength: 100, IsPrimaryKey: true},
{Name: "vector", DataType: FloatVector, Dim: 128}, // 行为特征向量
{Name: "student_id", DataType: VarChar, MaxLength: 100},
{Name: "behavior_type", DataType: VarChar, MaxLength: 50}, // preview/review/homework/mistake等
{Name: "lesson_id", DataType: VarChar, MaxLength: 100},
{Name: "duration", DataType: Int32}, // 行为时长（分钟）
{Name: "completion_rate", DataType: Double},
{Name: "quality_score", DataType: Double},
{Name: "timestamp", DataType: Int64},
]
EnableDynamicField: true
```

**索引**：HNSW (L2, M=16, efConstruction=200)

## 四、数据关系图

```
Student (学生)
├── attends → Course (参加课程)
├── studies → KnowledgePoint (学习知识点)
├── writes → Article (写文章)
├── answers → Question (答题)
├── chats → Conversation (会话)
├── belongs_to → FamilyMember (家庭成员)
├── friends_with → Student (同学关系)
├── likes → Hobby (兴趣爱好)
├── has_mbti → MBTITest (MBTI类型)
├── previews → Lesson (预习)
├── reviews → Lesson (复习)
├── completes_homework → Lesson (完成作业)
├── collects_mistake → Question (整理错题)
├── listens → Lesson (课堂听讲)
└── interacts → Lesson (课堂互动)

Lesson (课时)
├── contains → KnowledgePoint (包含知识点)
└── belongs_to → Course (属于课程)

KnowledgePoint (知识点)
└── prerequisite → KnowledgePoint (前置关系)
```

## 五、查询场景设计

### 5.1 基础身份信息查询

```ngql
// 获取学生基本信息
MATCH (s:student {id: "student_001"})
RETURN s.name, s.age, s.grade, s.gender, s.avatar_url, s.school, s.class_name;

// 获取学生形象特征（从会话中提取）
// 通过 Milvus 检索相关会话向量，提取形象描述
```

### 5.2 身体状况查询

```ngql
// 从会话中提取身体状况信息（通过向量检索）
// 查询生活习惯相关会话
MATCH (s:student {id: "student_001"})-[c:chats]->(conv:conversation)
WHERE conv.conversation_type == "首页"
RETURN conv.id;
// 然后通过 Milvus 检索这些会话的向量，查找身体状况相关内容
```

### 5.3 学习能力查询

```ngql
// 获取学生学习知识点情况
MATCH (s:student {id: "student_001"})-[st:studies]->(kp:knowledge_point)
RETURN kp.name, st.mastery_level, st.study_count, st.total_study_duration
ORDER BY st.mastery_level DESC;

// 获取学生答题正确率
MATCH (s:student {id: "student_001"})-[a:answers]->(q:question)
RETURN 
    COUNT(*) as total_questions,
    SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END) as correct_count,
    AVG(a.time_spent) as avg_time_spent;

// 通过 Milvus 检索学习行为向量，分析学习模式
```

### 5.4 学科能力查询

```ngql
// 获取学生在各学科的表现
MATCH (s:student {id: "student_001"})-[a:attends]->(c:course)
MATCH (s)-[st:studies]->(kp:knowledge_point)
WHERE kp.subject == c.subject
RETURN 
    c.subject,
    AVG(st.mastery_level) as avg_mastery,
    COUNT(DISTINCT kp.id) as knowledge_count;

// 获取学生文章评分趋势
MATCH (s:student {id: "student_001"})-[w:writes]->(art:article)
RETURN 
    art.lesson_id,
    w.score,
    w.write_time
ORDER BY w.write_time;
```

### 5.5 性格特征查询

```ngql
// 获取学生MBTI类型
MATCH (s:student {id: "student_001"})-[h:has_mbti]->(m:mbti_test)
RETURN m.mbti_type, m.e_score, m.i_score, m.s_score, m.n_score, 
       m.t_score, m.f_score, m.j_score, m.p_score
ORDER BY m.test_date DESC
LIMIT 1;

// 从会话中提取性格特征（通过向量检索）
```

### 5.6 学习习惯查询

```ngql
// 获取学生预习情况
MATCH (s:student {id: "student_001"})-[p:previews]->(l:lesson)
RETURN 
    COUNT(*) as preview_count,
    AVG(p.completion_rate) as avg_completion_rate,
    AVG(p.preview_duration) as avg_duration;

// 获取学生复习情况
MATCH (s:student {id: "student_001"})-[r:reviews]->(l:lesson)
RETURN 
    COUNT(*) as review_count,
    AVG(r.review_duration) as avg_duration,
    AVG(r.review_count) as avg_review_times;

// 获取学生作业完成情况
MATCH (s:student {id: "student_001"})-[h:completes_homework]->(l:lesson)
RETURN 
    COUNT(*) as homework_count,
    SUM(CASE WHEN h.is_on_time THEN 1 ELSE 0 END) as on_time_count,
    AVG(h.score) as avg_score,
    AVG(h.completion_time) as avg_completion_time;

// 获取学生错题整理情况
MATCH (s:student {id: "student_001"})-[cm:collects_mistake]->(q:question)
RETURN 
    COUNT(*) as mistake_count,
    SUM(CASE WHEN cm.is_mastered THEN 1 ELSE 0 END) as mastered_count,
    AVG(cm.review_count) as avg_review_count;
```

### 5.7 社会关系和背景查询

```ngql
// 获取学生家庭背景
MATCH (s:student {id: "student_001"})-[b:belongs_to]->(fm:family_member)
RETURN fm.name, fm.relationship, fm.education_level, fm.occupation, b.closeness;

// 获取学生社交网络
MATCH (s:student {id: "student_001"})-[f:friends_with]->(s2:student)
RETURN s2.name, s2.grade, f.friendship_level, f.since;
```

### 5.8 兴趣爱好查询

```ngql
// 获取学生兴趣爱好
MATCH (s:student {id: "student_001"})-[l:likes]->(h:hobby)
RETURN h.name, h.category, l.interest_level, l.since
ORDER BY l.interest_level DESC;
```

### 5.9 综合画像查询（结合向量检索）

```ngql
// 1. 从 Nebula 获取结构化数据
MATCH (s:student {id: "student_001"})
OPTIONAL MATCH (s)-[h:has_mbti]->(m:mbti_test)
OPTIONAL MATCH (s)-[st:studies]->(kp:knowledge_point)
OPTIONAL MATCH (s)-[a:attends]->(c:course)
RETURN 
    s.*,
    m.mbti_type,
    AVG(st.mastery_level) as avg_mastery,
    COLLECT(DISTINCT c.subject) as subjects;

// 2. 通过 Milvus 检索相关会话向量（语义搜索）
// 3. 通过 Milvus 检索文章向量（分析写作能力）
// 4. 通过 Milvus 检索学习行为向量（分析学习习惯）
// 5. 将所有数据整合后提供给大模型进行综合分析
```

## 六、RAG 数据召回策略

### 6.1 会话内容召回

```go
// 1. 通过 Nebula 获取学生会话ID列表
// 2. 使用 Milvus 向量检索，根据查询语义召回相关会话
// 3. 返回最相关的 N 条会话内容

query := "学生最近的学习状态如何？"
vector := embeddingModel.Encode(query)
results := milvus.Search("conversation_vectors", vector, topK = 10,
filter = "student_id == 'student_001'")
```

### 6.2 文章内容召回

```go
// 召回学生写作相关文章，用于分析写作能力
query := "学生的写作风格和表达能力"
vector := embeddingModel.Encode(query)
results := milvus.Search("article_vectors", vector, topK = 5,
filter = "student_id == 'student_001'")
```

### 6.3 学习行为召回

```go
// 召回相似的学习行为模式
query := "学生的学习习惯和时间管理"
vector := embeddingModel.Encode(query)
results := milvus.Search("learning_behavior_vectors", vector, topK = 20,
filter = "student_id == 'student_001'")
```

## 七、实现建议

1. **数据同步**：建立事件驱动机制，实时将业务数据写入 Nebula 和 Milvus
2. **向量生成**：使用中文 embedding 模型（如 text2vec-chinese）生成向量
3. **索引优化**：根据查询频率和模式优化 Nebula 和 Milvus 索引
4. **缓存策略**：对常用查询结果进行缓存，提高响应速度
5. **数据一致性**：确保 Nebula 和 Milvus 数据的一致性

## 八、下一步工作

1. 实现 Schema 初始化代码
2. 实现数据写入服务
3. 实现查询服务（Nebula + Milvus）
4. 实现 RAG 召回服务
5. 实现学生画像生成服务（整合数据 + 大模型）

