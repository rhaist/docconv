package docconv

import (
	"io"
	"os"
	"os/exec"
)

// Tidy attempts to tidy up XML.
// Errors & warnings are deliberately suppressed as underlying tools
// throw warnings very easily.
func Tidy(r io.Reader, xmlIn bool) ([]byte, error) {
	f, err := os.CreateTemp(os.TempDir(), "docconv")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	if _, err = io.Copy(f, r); err != nil {
		f.Close()
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}

	var output []byte
	if xmlIn {
		output, err = exec.Command("tidy", "-xml", "-numeric", "-asxml", "-quiet", "-utf8", f.Name()).Output()
	} else {
		output, err = exec.Command("tidy", "-numeric", "-asxml", "-quiet", "-utf8", f.Name()).Output()
	}

	if err != nil && err.Error() != "exit status 1" {
		return nil, err
	}
	return output, nil
}
