package filehelper

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetSHA256Sum(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGetSHA256Sum")
	sumText := []byte("hello world\n")
	filePath := filepath.Join(tmp, "test.file")
	err := ioutil.WriteFile(filePath, sumText, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	calcSha, err := GetSHA256Sum(filePath)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	if _, err := h.Write(sumText); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(h.Sum(nil), calcSha) {
		t.Fatal("bytes are not identical")
	}
}

func TestCopyFileContents(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyFileContents")
	sumText := []byte("hello world\n")
	filePath := filepath.Join(tmp, "test.file")
	err := ioutil.WriteFile(filePath, sumText, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	calcSha, err := GetSHA256Sum(filePath)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(tmp, "copy.file")
	if err := CopyFileContents(filePath, output); err != nil {
		t.Fatal(err)
	}
	copySha, err := GetSHA256Sum(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(copySha, calcSha) {
		t.Fatal("bytes are not identical")
	}
}
