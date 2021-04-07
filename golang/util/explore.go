package util

func GetCostByArea(area int) int {
	if area < 4 {
		return 1
	}
	if area < 8 {
		return 2
	}
	if area < 16 {
		return 3
	}
	if area < 32 {
		return 4
	}
	if area < 64 {
		return 5
	}
	if area < 128 {
		return 6
	}
	if area < 256 {
		return 7
	}
	if area < 512 {
		return 8
	}
	if area < 1024 {
		return 9
	}
	if area < 2048 {
		return 10
	}
	if area < 4096 {
		return 11
	}
	if area < 8192 {
		return 12
	}
	if area < 16384 {
		return 13
	}
	return 1
}

func GetCostByDepthInt(depth int) int {
	return 20 + (depth-1)*2
}

func GetCostByDepth(depth int) float32 {
	return 2 + float32(depth-1)*0.2
}
