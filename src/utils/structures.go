package utils

// Create a map from array with kf providing keys, values are array elements
func ArrayToMap[T, K comparable](ts []T, kf func(T) K) map[K]T {
	result := make(map[K]T)
	for _, t := range ts {
		result[kf(t)] = t
	}
	return result
}

// Create a map from array with kf providing keys, values are pointers to array elements
func ArrayToPtrMap[T, K comparable](ts []T, kf func(T) K) map[K]*T {
	result := make(map[K]*T)
	for _, t := range ts {
		result[kf(t)] = &t
	}
	return result
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
