package cache

// caching to get api results faster

type NameCache interface {
	Set(key string, value []byte)
	Get(key string) []byte
}
