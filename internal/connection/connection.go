package connection

type Connection interface {
    Write([]byte) error
    Close() error
}