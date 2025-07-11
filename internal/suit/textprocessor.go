package suit

import (
	"log"
	"runtime"
	"textool/pkg/fileutil"
)

const (
	DefaultChunkSize = 1500
	DefaultOverlap   = 200
)

var DefaultMaxWorkers = runtime.NumCPU() * 2 // 需导入"runtime"
// MdiftConfig 主处理结构体，负责管理文本处理的核心配置和资源
// 字段说明:
//   - apiKey: API访问密钥，用于第三方服务认证(长度应大于32字符)
//   - chunkSize: 文本分块大小(单位：字符)，建议值500-2000
//   - overlap: 块间重叠字符数，用于保持上下文连贯性
//   - maxWorkers: 最大并行工作协程数，受CPU核心数限制
//   - cacheDir: 缓存目录路径，用于存储中间处理结果
//   - logger: 结构化日志记录器，用于系统运行状态跟踪
type MdiftConfig struct {
	apiKey     string
	chunkSize  int
	overlap    int
	maxWorkers int
}

// NewMdiftConfig 创建并初始化文本处理器实例
// 参数:
//   - apiKey: 必填，API服务认证密钥
//   - chunkSize: 文本分块大小，<=0时使用默认值1500
//   - overlap: 块间重叠字符数，<=0时使用默认值200
//   - maxWorkers: 最大并行数，<=0时使用CPU核心数*2
//
// 返回:
//   - 初始化完成的MdiftConfig引用
func NewMdiftConfig(apiKey string, chunkSize, overlap, maxWorkers int) *MdiftConfig {
	// 设置默认值
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if overlap <= 0 {
		overlap = DefaultOverlap
	}
	if maxWorkers <= 0 {
		maxWorkers = DefaultMaxWorkers
	}

	return &MdiftConfig{
		apiKey:     apiKey,
		chunkSize:  chunkSize,
		overlap:    overlap,
		maxWorkers: maxWorkers,
	}
}

// ProcessFile 处理文件的主入口
func (tp *MdiftConfig) ProcessFile(inputPath, outputPath string) {
	fullText := fileutil.ReadFile(inputPath)
	log.Println("开始处理文件：", inputPath)

	chunks := splitText(fullText, tp.chunkSize)
	log.Println("已分块：", len(chunks))

	results := tp.parallelProcessing(chunks)
	log.Println("已处理：", len(results))

	finalMD := postProcess(joinResults(results))
	log.Println("已生成结果：", outputPath)

	fileutil.WriteFile(outputPath, finalMD)
	log.Println("处理完成")
}
