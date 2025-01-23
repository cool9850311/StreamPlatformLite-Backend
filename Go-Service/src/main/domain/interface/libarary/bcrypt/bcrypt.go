package bcrypt

type BcryptGenerator interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}
