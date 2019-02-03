package storage

import (
	"fmt"
	"testing"
)

const TestObjectContent = "Some Test Content"

func TestNewS3Client(t *testing.T) {
	_, err := newS3Client()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestNewS3Object(t *testing.T) {
	o, err := NewS3Object("someFile.zip")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	err = o.Write([]byte(TestObjectContent))
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	buf, err := o.Read()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if string(buf) != TestObjectContent {
		t.Errorf("Expected object to be '%s', was '%s'", TestObjectContent, string(buf))
		t.Fail()
	}
}