package agentdctl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	maxLength            = 512
	ErrInterrupted       = errors.New("interrupted")
	ErrMaxLengthExceeded = fmt.Errorf("maximum byte limit (%v) exceeded", maxLength)
)

func input(def string, password bool) string {
	var buff []byte

	if password {
		buff, _ = getPwd(true)
	} else {
		buff, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	if str := strings.TrimSpace(string(buff)); str != "" {
		return str
	}

	return def
}

func getch(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	if n, err := r.Read(buf); n == 0 || err != nil {
		if err != nil {
			return 0, err
		}
		return 0, io.EOF
	}
	return buf[0], nil
}

func getPwd(masked bool) ([]byte, error) {
	var err error
	var pass, bs, mask []byte

	if masked {
		bs = []byte("\b \b")
		mask = []byte("*")
	}

	r := os.Stdin
	w := os.Stdout

	rfd := int(r.Fd())

	if terminal.IsTerminal(rfd) {
		if state, err := terminal.MakeRaw(rfd); err != nil {
			return pass, err
		} else {
			defer func() {
				terminal.Restore(rfd, state)
				w.Write([]byte("\n"))
			}()
		}
	}

	var counter int
	for counter = 0; counter <= maxLength; counter++ {
		if v, e := getch(r); e != nil {
			err = e
			break
		} else if v == 127 || v == 8 {
			if l := len(pass); l > 0 {
				pass = pass[:l-1]
				fmt.Fprint(w, string(bs))
			}
		} else if v == 13 || v == 10 {
			break
		} else if v == 3 {
			err = ErrInterrupted
			break
		} else if v != 0 {
			pass = append(pass, v)
			fmt.Fprint(w, string(mask))
		}
	}

	if counter > maxLength {
		err = ErrMaxLengthExceeded
	}

	return pass, err
}
