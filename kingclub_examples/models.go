package kingclub

import "time"

// ==================== Nebula Graph 数据模型 ====================

// Student 学生
type Student struct {
	ID             string
	Name           string
	Age            int
	Grade          string
	Gender         string
	AvatarURL      string
	Phone          string
	Email          string
	School         string
	ClassName      string
	EnrollmentDate time.Time
	CreateAt       time.Time
	UpdateAt       time.Time
}

// Course 课程
type Course struct {
	ID          string
	Name        string
	Subject     string
	Level       string
	Description string
	Duration    int // 分钟
	CreateAt    time.Time
}

// Lesson 课时
type Lesson struct {
	ID         string
	CourseID   string
	Title      string
	LessonType string // 写作课、阅读课等
	StartTime  time.Time
	EndTime    time.Time
	CreateAt   time.Time
}

// KnowledgePoint 知识点
type KnowledgePoint struct {
	ID          string
	Name        string
	Subject     string
	Difficulty  int // 1-5
	Description string
	CreateAt    time.Time
}

// Article 文章
type Article struct {
	ID        string
	StudentID string
	LessonID  string
	Title     string
	Content   string
	WordCount int
	Score     float64
	Feedback  string
	CreateAt  time.Time
}

// Question 题目
type Question struct {
	ID            string
	LessonID      string
	QuestionType  string // 选择题、填空题等
	Content       string
	CorrectAnswer string
	Difficulty    int
	CreateAt      time.Time
}

// Conversation 会话
type Conversation struct {
	ID               string
	StudentID        string
	LessonID         string // 可为空（首页会话）
	ConversationType string // 首页、课中、课后
	Summary          string
	MessageCount     int
	StartTime        time.Time
	EndTime          time.Time
	CreateAt         time.Time
}

// FamilyMember 家庭成员
type FamilyMember struct {
	ID             string
	Name           string
	Relationship   string // 父亲、母亲等
	EducationLevel string
	Occupation     string
	CreateAt       time.Time
}

// Teacher 老师
type Teacher struct {
	ID              string
	Name            string
	Subject         string
	ExperienceYears int
	CreateAt        time.Time
}

// Hobby 兴趣爱好
type Hobby struct {
	ID       string
	Name     string
	Category string // 运动、艺术、科技等
	CreateAt time.Time
}

// MBTITest MBTI测试
type MBTITest struct {
	ID        string
	StudentID string
	MBTIType  string // INTJ、ENFP等
	EScore    int
	IScore    int
	SScore    int
	NScore    int
	TScore    int
	FScore    int
	JScore    int
	PScore    int
	TestDate  time.Time
	CreateAt  time.Time
}

// ==================== 边（Edge）数据模型 ====================

// AttendsEdge 学生参加课程
type AttendsEdge struct {
	StudentID  string
	CourseID   string
	EnrollTime time.Time
	Progress   float64 // 0-1
	Status     string  // 进行中、已完成、已退课
}

// StudiesEdge 学生学习知识点
type StudiesEdge struct {
	StudentID          string
	KnowledgeID        string
	MasteryLevel       int // 1-5
	StudyCount         int
	LastStudyTime      time.Time
	TotalStudyDuration int // 分钟
}

// WritesEdge 学生写文章
type WritesEdge struct {
	StudentID string
	ArticleID string
	LessonID  string
	WriteTime time.Time
	WordCount int
	Score     float64
}

// AnswersEdge 学生答题
type AnswersEdge struct {
	StudentID     string
	QuestionID    string
	LessonID      string
	AnswerContent string
	IsCorrect     bool
	AnswerTime    time.Time
	TimeSpent     int // 秒
}

// ChatsEdge 学生会话
type ChatsEdge struct {
	StudentID       string
	ConversationID  string
	LessonID        string
	MessageCount    int
	AvgResponseTime float64 // 秒
	EngagementScore float64 // 0-1
}

// BelongsToEdge 学生属于家庭
type BelongsToEdge struct {
	StudentID      string
	FamilyMemberID string
	Relationship   string
	Closeness      int // 1-5
}

// FriendsWithEdge 学生与同学关系
type FriendsWithEdge struct {
	StudentID1      string
	StudentID2      string
	FriendshipLevel int // 1-5
	Since           time.Time
}

// LikesEdge 学生喜欢兴趣
type LikesEdge struct {
	StudentID     string
	HobbyID       string
	InterestLevel int // 1-5
	Since         time.Time
}

// HasMBTIEdge 学生有MBTI类型
type HasMBTIEdge struct {
	StudentID  string
	MBTITestID string
	MBTIType   string
	TestDate   time.Time
}

// PreviewsEdge 学生预习
type PreviewsEdge struct {
	StudentID       string
	LessonID        string
	PreviewTime     time.Time
	PreviewDuration int     // 分钟
	CompletionRate  float64 // 0-1
}

// ReviewsEdge 学生复习
type ReviewsEdge struct {
	StudentID      string
	LessonID       string
	ReviewTime     time.Time
	ReviewDuration int // 分钟
	ReviewCount    int
}

// CompletesHomeworkEdge 学生完成作业
type CompletesHomeworkEdge struct {
	StudentID      string
	LessonID       string
	HomeworkID     string
	SubmitTime     time.Time
	CompletionTime int // 分钟
	Score          float64
	IsOnTime       bool
}

// CollectsMistakeEdge 学生整理错题
type CollectsMistakeEdge struct {
	StudentID   string
	QuestionID  string
	CollectTime time.Time
	ReviewCount int
	IsMastered  bool
}

// ListensEdge 学生课堂听讲
type ListensEdge struct {
	StudentID        string
	LessonID         string
	AttentionScore   float64 // 0-1
	InteractionCount int
	FocusDuration    int // 分钟
}

// InteractsEdge 学生课堂互动
type InteractsEdge struct {
	StudentID       string
	LessonID        string
	InteractionType string // 提问、回答、讨论等
	InteractionTime time.Time
	QualityScore    float64 // 0-1
}

// ==================== Milvus 向量数据模型 ====================

// ConversationVector 会话向量
type ConversationVector struct {
	ID               string
	Vector           []float32
	ConversationID   string
	StudentID        string
	LessonID         string
	MessageType      string // user/assistant
	Content          string
	Timestamp        int64
	ConversationType string
	Metadata         map[string]interface{}
}

// ArticleVector 文章向量
type ArticleVector struct {
	ID        string
	Vector    []float32
	ArticleID string
	StudentID string
	LessonID  string
	Title     string
	Content   string
	WordCount int
	Score     float64
	Timestamp int64
	Metadata  map[string]interface{}
}

// QuestionAnswerVector 问答向量
type QuestionAnswerVector struct {
	ID              string
	Vector          []float32
	QuestionID      string
	StudentID       string
	LessonID        string
	QuestionContent string
	AnswerContent   string
	IsCorrect       bool
	TimeSpent       int
	Timestamp       int64
	Metadata        map[string]interface{}
}

// LearningBehaviorVector 学习行为向量
type LearningBehaviorVector struct {
	ID             string
	Vector         []float32
	StudentID      string
	BehaviorType   string // preview/review/homework/mistake等
	LessonID       string
	Duration       int // 分钟
	CompletionRate float64
	QualityScore   float64
	Timestamp      int64
	Metadata       map[string]interface{}
}

// ==================== 查询结果模型 ====================

// StudentProfile 学生画像
type StudentProfile struct {
	// 基础身份信息
	BasicInfo *StudentBasicInfo

	// 身体状况
	PhysicalCondition *PhysicalCondition

	// 学习能力
	LearningAbility *LearningAbility

	// 学科能力
	SubjectAbility *SubjectAbility

	// 性格特征
	Personality *Personality

	// 学习习惯
	LearningHabits *LearningHabits

	// 社会关系和背景
	SocialBackground *SocialBackground

	// 兴趣爱好
	Interests []*Interest
}

// StudentBasicInfo 基础身份信息
type StudentBasicInfo struct {
	ID            string
	Name          string
	Age           int
	Grade         string
	Gender        string
	AvatarURL     string
	School        string
	ClassName     string
	ImageFeatures string // 形象特征（从会话中提取）
}

// PhysicalCondition 身体状况
type PhysicalCondition struct {
	PhysicalAbility string // 身体能力描述
	LivingHabits    string // 生活习惯描述
	HealthStatus    string // 健康状况
}

// LearningAbility 学习能力
type LearningAbility struct {
	CognitiveLevel     string  // 认知水平
	ThinkingMode       string  // 思维模式
	KnowledgeStructure string  // 知识结构
	AverageMastery     float64 // 平均掌握度
	TotalKnowledge     int     // 总知识点数
}

// SubjectAbility 学科能力
type SubjectAbility struct {
	Subjects map[string]*SubjectPerformance // 学科 -> 表现
}

// SubjectPerformance 学科表现
type SubjectPerformance struct {
	Subject        string
	AverageScore   float64
	KnowledgeCount int
	AverageMastery float64
	ArticleScores  []float64 // 文章评分趋势
}

// Personality 性格特征
type Personality struct {
	MBTIType string
	EScore   int
	IScore   int
	SScore   int
	NScore   int
	TScore   int
	FScore   int
	JScore   int
	PScore   int
	Traits   []string // 性格特质列表
}

// LearningHabits 学习习惯
type LearningHabits struct {
	Preview           *HabitStats // 预习
	Review            *HabitStats // 复习
	Homework          *HabitStats // 作业
	MistakeCollection *HabitStats // 错题整理
	TimeManagement    string      // 时间管理
}

// HabitStats 习惯统计
type HabitStats struct {
	Count           int
	AverageRate     float64
	AverageDuration int
	Consistency     string // 一致性描述
}

// SocialBackground 社会关系和背景
type SocialBackground struct {
	FamilyMembers []*FamilyMemberInfo
	Friends       []*FriendInfo
}

// FamilyMemberInfo 家庭成员信息
type FamilyMemberInfo struct {
	Name           string
	Relationship   string
	EducationLevel string
	Occupation     string
	Closeness      int
}

// FriendInfo 朋友信息
type FriendInfo struct {
	Name            string
	Grade           string
	FriendshipLevel int
	Since           time.Time
}

// Interest 兴趣爱好
type Interest struct {
	Name          string
	Category      string
	InterestLevel int
	Since         time.Time
}
