package container

import (
	"testing"
)

func TestPathIsExist(t *testing.T) {
	existPath := "/root/mnt"
	notExistPath := "/root/123"

	exist, err := PathIsExist(existPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v exist: %v\n", existPath, exist)

	exist, err = PathIsExist(notExistPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v exist: %v\n", notExistPath, exist)
}

func TestCreateWriteLayer(t *testing.T) {
	CreateWriteLayer("/root/")
}

func TestCreateOrClear(t *testing.T) {
	existPath := "/root/busybox"
	notExistPath := "/root/123"

	if err := CreateOrClear(existPath); err != nil {
		panic(err)
	}
	if err := CreateOrClear(notExistPath); err != nil {
		panic(err)
	}
}