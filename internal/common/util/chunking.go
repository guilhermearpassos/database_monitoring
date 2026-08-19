package util

// PackByMaxBytes groups items into contiguous chunks such that the size
// of a chunk, as reported by sizeWith when tentatively appending the next item,
// does not exceed maxBytes.
//
// - items are packed greedily in order; boundaries are deterministic.
// - if maxBytes <= 0, a single chunk with all items is returned.
// - if an individual item would exceed maxBytes even in an empty chunk,
//   it is placed alone in its own chunk.
//
// The sizeWith function receives the current chunk (may be empty) and the next
// item and must return the uncompressed byte size of the whole chunk if the
// next item were appended. Callers can reuse a preallocated probe object to
// compute this efficiently.
func PackByMaxBytes[T any](items []T, maxBytes int, sizeWith func(cur []T, next T) int) [][]T {
	if len(items) == 0 {
		return [][]T{}
	}
	if maxBytes <= 0 {
		// Single chunk containing everything
		chunk := make([]T, len(items))
		copy(chunk, items)
		return [][]T{chunk}
	}
	res := make([][]T, 0)
	cur := make([]T, 0, 256)
	flush := func() {
		if len(cur) == 0 {
			return
		}
		chunk := make([]T, len(cur))
		copy(chunk, cur)
		res = append(res, chunk)
		cur = cur[:0]
	}
	for _, it := range items {
		sz := sizeWith(cur, it)
		if sz <= maxBytes {
			cur = append(cur, it)
			continue
		}
		// If current is empty, force single-item chunk
		if len(cur) == 0 {
			res = append(res, []T{it})
			cur = cur[:0]
			continue
		}
		// Otherwise flush current and start a new chunk with it
		flush()
		cur = append(cur, it)
	}
	flush()
	return res
}
