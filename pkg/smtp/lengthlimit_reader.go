package smtp

import (
	"errors"
	"io"
)

var ErrTooLongLine = errors.New("smtp: too long a line in input stream")

type lineLimitReader struct {
	R         io.Reader
	LineLimit int

	curLineLength int
}

func (r *lineLimitReader) Read(b []byte) (int, error) {
	if r.curLineLength > r.LineLimit && r.LineLimit > 0 {
		return 0, ErrTooLongLine
	}

	n, err := r.R.Read(b)
	if err != nil {
		return n, err
	}

	if r.LineLimit == 0 {
		return n, nil
	}

	for _, chr := range b[:n] {
		if chr == '\n' {
			r.curLineLength = 0
		}
		r.curLineLength++

		if r.curLineLength > r.LineLimit {
			return 0, ErrTooLongLine
		}
	}

	return n, nil
}
