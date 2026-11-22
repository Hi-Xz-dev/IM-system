package connection

import (
	"bufio"
	"fmt"
	"io"
)

type TCPReader struct {
	scanner *bufio.Scanner
}

func NewTCPReader(conn io.Reader) *TCPReader {
	scanner := bufio.NewScanner(conn)
	// 额外预留 CRLF 的两个字节，行结束符不计入消息长度。
	scanner.Buffer(nil, MaxMessageBytes+2)
	return &TCPReader{scanner: scanner}
}

func (t *TCPReader) Read() (string, error) {

	ok := t.scanner.Scan()
	// Scanner 在底层读取超时后仍可能交出已经读到的半行。
	if err := t.scanner.Err(); err != nil {
		return "", err
	}
	if !ok {
		return "", io.EOF
	}
	if len(t.scanner.Bytes()) > MaxMessageBytes {
		return "", fmt.Errorf("tcp message exceeds %d bytes", MaxMessageBytes)
	}
	return t.scanner.Text(), nil
}
