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
	err := ioutil.WriteFile(filePath, sumText, 777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	calcSha, err := GetSHA256Sum(filePath)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	h.Write(sumText)
	if bytes.Compare(h.Sum(nil), calcSha) != 0 {
		t.Fatal("bytes are not identical")
	}
}
