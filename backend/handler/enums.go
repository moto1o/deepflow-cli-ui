package handler

// 枚举映射，对照 cli/ctl/common/const.go 的 stringer line-comment。

var vtapStateName = map[int]string{
	0: "LOST", // VTAP_STATE_NOT_CONNECTED
	1: "NORMAL",
	2: "DISABLE",
	3: "PENDING",
}

var vtapTypeName = map[int]string{
	1:  "KVM",
	2:  "ESXI",
	3:  "CHOST_VM", // WORKLOAD_V
	5:  "CHOST_BM", // WORKLOAD_P
	6:  "DEDICATED",
	7:  "K8S_BM",    // POD_HOST
	8:  "K8S_VM",    // POD_VM
	9:  "TUN_DECAP", // TUNNEL_DECAPSULATION
	10: "HYPER_V",
	12: "K8S_SIDECAR",
}

var vtapExceptionName = map[int64]string{
	0x10000000: "LICENSE_NOT_ENGOUTH",
	0x40000000: "ALLOC_ANALYZER_FAILED",
	0x80000000: "ALLOC_CONTROLLER_FAILED",
}

// tap_mode 名称 → int（对照 GetVtapTapModeByName）
var tapModeValue = map[string]int{
	"local":     0,
	"mirror":    1,
	"dedicated": 2,
}

func vtapStateStr(v int) string {
	if s, ok := vtapStateName[v]; ok {
		return s
	}
	return "UNKNOWN"
}

func vtapTypeStr(v int) string {
	if s, ok := vtapTypeName[v]; ok {
		return s
	}
	return "UNKNOWN"
}

func vtapExceptionStr(v int64) string {
	if s, ok := vtapExceptionName[v]; ok {
		return s
	}
	return ""
}
