//go:build !linux && !darwin

package tools

func init() {
	getMemoryInfo = func() string { return "" }
	getDiskInfo = func() string { return "" }
}
