package monitoring

type testParameter struct {
	workDir       string
	command       string
	args          []string
	runningMode   int32
	parallelCount int32
}

func (o *testParameter) WorkDir() string {
	return o.workDir
}

func (o *testParameter) Command() string {
	return o.command
}

func (o *testParameter) ToArgs() []string {
	return o.args
}

func (o *testParameter) RunningMode() int32 {
	return o.runningMode
}

func (o *testParameter) ParallelCount() int32 {
	return o.parallelCount
}

func (o *testParameter) StdErrIsOk() bool {
	return false
}

func (o *testParameter) CheckStartLine() func(string) bool {
	return nil
}