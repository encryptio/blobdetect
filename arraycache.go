package blobdetect

// By caching the indicies int array, we have zero allocations in the good
// data path. This increases performance by 2-5% depending on the image.

var arrayCache = make(chan []int, 10)

func getIntArray(size int) []int {
	select {
	case b := <-arrayCache:
		if cap(b) >= size {
			return b[0:size]
		}
	default:
	}
	return make([]int, size)
}

func releaseIntArray(arr []int) {
	select {
	case arrayCache <- arr:
	default:
	}
}
