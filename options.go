package shp

// ReaderOption 定义读取器选项
type ReaderOption func(*ReaderConfig)

// ReaderConfig 读取器配置
type ReaderConfig struct {
	// IgnoreCorruptedShapes 是否忽略损坏的形状
	IgnoreCorruptedShapes bool
	// MaxMemoryUsage 最大内存使用量(字节)
	MaxMemoryUsage int64
	// EnableBuffering 是否启用缓冲
	EnableBuffering bool
	// BufferSize 缓冲区大小
	BufferSize int
	// Debug 是否启用调试输出
	Debug bool
}

// DefaultReaderConfig 默认读取器配置
func DefaultReaderConfig() *ReaderConfig {
	return &ReaderConfig{
		IgnoreCorruptedShapes: false,
		MaxMemoryUsage:        100 * 1024 * 1024, // 100MB
		EnableBuffering:       true,
		BufferSize:            64 * 1024, // 64KB
		Debug:                 false,     // 默认关闭调试输出
	}
}

// WithIgnoreCorruptedShapes 设置是否忽略损坏的形状
func WithIgnoreCorruptedShapes(ignore bool) ReaderOption {
	return func(config *ReaderConfig) {
		config.IgnoreCorruptedShapes = ignore
	}
}

// WithMaxMemoryUsage 设置最大内存使用量
func WithMaxMemoryUsage(size int64) ReaderOption {
	return func(config *ReaderConfig) {
		config.MaxMemoryUsage = size
	}
}

// WithBuffering 设置缓冲选项
func WithBuffering(enabled bool, size int) ReaderOption {
	return func(config *ReaderConfig) {
		config.EnableBuffering = enabled
		config.BufferSize = size
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) ReaderOption {
	return func(config *ReaderConfig) {
		config.Debug = debug
	}
}

// WriterOption 定义写入器选项
type WriterOption func(*WriterConfig)

// WriterConfig 写入器配置
type WriterConfig struct {
	// CompressionLevel 压缩级别 (0-9)
	CompressionLevel int
	// EnableValidation 是否启用数据验证
	EnableValidation bool
	// BufferSize 缓冲区大小
	BufferSize int
	// EnableSync 是否在每次写入后同步到磁盘
	EnableSync bool
}

// DefaultWriterConfig 默认写入器配置
func DefaultWriterConfig() *WriterConfig {
	return &WriterConfig{
		CompressionLevel: 0,
		EnableValidation: true,
		BufferSize:       64 * 1024, // 64KB
		EnableSync:       false,
	}
}

// WithCompressionLevel 设置压缩级别
func WithCompressionLevel(level int) WriterOption {
	return func(config *WriterConfig) {
		if level >= 0 && level <= 9 {
			config.CompressionLevel = level
		}
	}
}

// WithValidation 设置数据验证选项
func WithValidation(enabled bool) WriterOption {
	return func(config *WriterConfig) {
		config.EnableValidation = enabled
	}
}

// WithWriterBuffering 设置写入器缓冲选项
func WithWriterBuffering(size int) WriterOption {
	return func(config *WriterConfig) {
		config.BufferSize = size
	}
}

// WithSync 设置同步选项
func WithSync(enabled bool) WriterOption {
	return func(config *WriterConfig) {
		config.EnableSync = enabled
	}
}
