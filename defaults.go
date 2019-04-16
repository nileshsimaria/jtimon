package main

const (
	// DefaultGRPCWindowSize is the default GRPC Window Size
	DefaultGRPCWindowSize = 1048576

	// DefaultIDBBatchSize to use if user has not provided in the config
	DefaultIDBBatchSize = 1024 * 100
	//DefaultIDBBatchFreq is 2 seconds
	DefaultIDBBatchFreq = 2000
	//DefaultIDBAccumulatorFreq is 2 seconds
	DefaultIDBAccumulatorFreq = 2000
	//DefaultIDBTimeout is 30 seconds
	DefaultIDBTimeout = 30

	// MatchExpressionXpath is for the pattern matching the xpath and key-value pairs
	MatchExpressionXpath = "\\/([^\\/]*)\\[(.*?)+?(?:\\])"
	// MatchExpressionKey is for pattern matching the single and multiple key value pairs
	MatchExpressionKey = "([A-Za-z0-9-/]*)=(.*?)?(?: and |$)+"
)
