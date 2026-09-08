package shell

import "io"

type interceptor struct {
	writer   io.Writer
	lastByte byte
}

// interceptor wraps io writer and tracks last byte written its purpose is to detect if a newline should be inserted
func (i *interceptor) Write(p []byte) (int, error) {
	for idx, b := range p {
		if idx == len(p)-1 {
			i.lastByte = b
		}
	}
	return i.writer.Write(p)
}

func (i *interceptor) LastByteNewLine() bool {
	return i.lastByte == '\n'
}

func (i *interceptor) LastByteTab() bool {
	return i.lastByte == '\t'
}
