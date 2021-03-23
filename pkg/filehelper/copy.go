package filehelper

import copy2 "github.com/otiai10/copy"

func Copy(src, dst string) error {
	return copy2.Copy(src, dst)
}
