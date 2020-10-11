package graph

// This file is taken from https://github.com/shawnsmithdev/zermelo/blob/v1.0.3/zuint32/zuint32.go
//
// github.com/shawnsmithdev/zermelo is licensed under MIT License and Copyright (c) 2014 Shawn Smith
//
// It is modified for this project to use Node instead of uint32

// sortNodesBYOB sorts x using a Radix sort, using supplied buffer space. Panics if
// len(x) does not equal len(buffer). Uses radix sort even on small slices.
func sortNodesBYOB(x, buffer []Node) {
	const radix uint = 8
	const bitSize uint = 32
	if len(x) > len(buffer) {
		panic("Buffer too small")
	}
	if len(x) < 2 {
		return
	}

	from := x
	to := buffer[:len(x)]

	for keyOffset := uint(0); keyOffset < bitSize; keyOffset += radix {
		var offset [256]int // Keep track of where room is made for byte groups in the buffer
		sorted := false
		prev := Node(0)

		for _, elem := range from {
			// For each elem to sort, fetch the byte at current radix
			key := uint8(elem >> keyOffset)
			// inc count of bytes of this type
			offset[key]++

			if sorted { // Detect sorted
				sorted = elem >= prev
				prev = elem
			}
		}

		if sorted { // Short-circuit sorted
			if (keyOffset/radix)%2 == 1 {
				copy(to, from)
			}
			return
		}

		// Find target bucket offsets
		watermark := 0
		//nolint:gocritic // rangeExprCopy appears to be a false positive
		for i, count := range offset {
			offset[i] = watermark
			watermark += count
		}

		// Swap values between the buffers by radix
		for _, elem := range from {
			key := uint8(elem >> keyOffset) // Get the byte of each element at the radix
			to[offset[key]] = elem          // Copy the element depending on byte offsets
			offset[key]++                   // One less space, move the offset
		}

		// Reverse buffers on each pass
		from, to = to, from
	}
}
