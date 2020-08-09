package store

// Store ...
type Store interface {
	Token() TokenRepository
}
