package archiver

import (
	"compress/flate"
	"github.com/mholt/archiver"
)

func Zip(input string, output string) error {

	z := archiver.Zip{
		CompressionLevel:       flate.DefaultCompression,
		MkdirAll:               true,
		SelectiveCompression:   true,
		ContinueOnError:        false,
		OverwriteExisting:      false,
		ImplicitTopLevelFolder: false,
	}

	return z.Archive([]string{input}, output)
}
