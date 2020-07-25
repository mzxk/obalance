package obalance

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
)

func join(s ...string) string {
	if len(s) == 0 {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteString(s[0])
	for i := 1; i < len(s); i++ {
		buf.WriteString(".")
		buf.WriteString(s[i])
	}
	return buf.String()
}
func s2f(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0.0
	}
	return f
}

//s2s 格式化float64格式的string类型到10位小数
func s2s(s string) string {
	return fmt.Sprintf("%.10f", s2f(s))
}
