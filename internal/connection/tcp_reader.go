package connection

import (
	"bufio"
	"io"
	"net"
)

type TCPReader struct {
	scanner *bufio.Scanner
}

func NewTCPReader(conn net.Conn) *TCPReader {

	return &TCPReader{
		scanner:bufio.NewScanner(conn),
	}
} 

func (t *TCPReader) Read() (string, error){

	if !t.scanner.Scan(){

		if err := t.scanner.Err(); err != nil{
			return "", err
		}
		return "", io.EOF
	}
	return t.scanner.Text(), nil
}