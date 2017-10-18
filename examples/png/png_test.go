package png

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestFixChecksums(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/gopher.png")
	if err != nil {
		t.Fatal(err)
	}
	fixedData, err := fixChecksums(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fixedData, data) {
		t.Error("fixChecksums modified a non-corrupt png")
	}
}
