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

const (
	FunctionDescriptionTag = "Function-Description"
	FunctionNameTag        = "Function-Name"
	FunctionHandlerTag     = "Function-Handler"
	FunctionRuntimeTag     = "Function-Runtime"
	FunctionMemorySizeTag  = "Function-Memory-Size"
	FunctionTimeoutTag     = "Function-Timeout"
	FunctionAliasTag       = "Function-Alias"
)
