package repository

type Store interface {
	Set(key, value string) error
	Get(key string) (string, error)
}
