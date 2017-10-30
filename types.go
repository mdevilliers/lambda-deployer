package deployer

// FunctionMetadata collects the definable parameters
// for a lambda function
type FunctionMetadata struct {
	Description  string
	FunctionName string
	Handler      string
	Runtime      string
	MemorySize   int64
	Timeout      int64
	Alias        string
	EnvVars      map[string]interface{}
}
