package human

import (
	"bytes"
	"os/user"
	"path/filepath"
)

// Path represents a path on the file system.
//
// The type interprets the special prefix "~/" as representing the home
// directory of the user that the program is running as.
type Path string

func (p *Path) UnmarshalText(b []byte) error {
	switch {
	case bytes.HasPrefix(b, []byte{'~', filepath.Separator}):
		u, err := user.Current()
		if err != nil {
			return err
		}
		*p = Path(filepath.Join(u.HomeDir, string(b[2:])))
	default:
		*p = Path(b)
	}
	return nil
}
