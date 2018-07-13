package util

import (
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

// Download downloads an attachment from a given URL to a destination file
func Download(url, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "couldn't creating destination file %s for attachment", dest)
	}
	defer out.Close()

	rsp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "couldn't download attachment from %s", url)
	}
	defer rsp.Body.Close()

	_, err = io.Copy(out, rsp.Body)
	if err != nil {
		return errors.Wrapf(err, "couldn't write attachment into destination file")
	}

	return nil
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "couldn't copy file "+src+": couldn't open source file")
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return errors.Wrap(err, "couldn't copy file "+src+": couldn't create destination file "+dst)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.Wrap(err, "couldn't copy file "+src+" to "+dst)
	}

	err = out.Close()
	if err != nil {
		return errors.Wrap(err, "couldn't copy file "+src+" to "+dst+": couldn't close destination file")
	}

	err = os.Chmod(dst, 0755)
	if err != nil {
		return errors.Wrap(err, "couldn't make destination file "+dst+" executable")
	}

	return nil
}
