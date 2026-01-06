package pkg

type Repository interface {
	SetMessageID(key, value string) error
	GetMessageID(key string) (string, error)
	DeleteMessageID(key string) error
}
