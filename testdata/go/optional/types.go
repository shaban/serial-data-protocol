package optional

// Request Basic struct with optional metadata (RC Feature 1 test)
type Request struct {
	// Id Request ID (always present)
	Id uint32
	// Metadata Optional user metadata
	Metadata *Metadata
}

// Metadata User metadata
type Metadata struct {
	// UserId User ID
	UserId uint64
	// Username Username
	Username string
}

// Config Complex example with nested optional fields
type Config struct {
	// Name Configuration name
	Name string
	// Database Optional database settings
	Database *DatabaseConfig
	// Cache Optional cache settings
	Cache *CacheConfig
}

// DatabaseConfig Database configuration
type DatabaseConfig struct {
	// Host Database host
	Host string
	// Port Database port
	Port uint16
}

// CacheConfig Cache configuration
type CacheConfig struct {
	// SizeMb Cache size in MB
	SizeMb uint32
	// TtlSeconds Time to live in seconds
	TtlSeconds uint32
}

// Document Example with optional array (wrapped struct containing array)
type Document struct {
	// Id Document ID
	Id uint32
	// Tags Optional tags wrapper
	Tags *TagList
}

// TagList Wrapper for tag array
type TagList struct {
	// Items Array of tags
	Items []string
}
