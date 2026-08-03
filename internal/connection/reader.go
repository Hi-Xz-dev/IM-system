package connection

type Reader interface {
	Read() (string, error)
}
