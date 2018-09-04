package stressng

type Param struct {
	Stressor      string
	StressorCount int64
	Timeout       int64
	TimeoutUnit   string
	Abort         bool
	Metrics       bool
}
