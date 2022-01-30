package container

import (
	"testing"
)

func TestPathIsExist(t *testing.T) {
	existPath := "/root/mnt"
	notExistPath := "/root/123"

	exist, err := pathIsExist(existPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v exist: %v\n", existPath, exist)

	exist, err = pathIsExist(notExistPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v exist: %v\n", notExistPath, exist)
}

func TestCreateWriteLayer(t *testing.T) {
	CreateWriteLayer("/root/")
}