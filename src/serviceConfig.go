package main

import "runtime"

//程序可以使用CPU核心数量
func useableCPUNum() int {
	u := configCPUS()
	if u == 0 {
		u = runtime.NumCPU()
		return u
	}
	return u
}

// //剔除不需要分析的账号
// func needExcludeUID() {

// }
