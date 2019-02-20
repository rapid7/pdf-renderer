package storage

import (
	"fmt"
	"testing"
	"os"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

const TestObjectContent = "Some Test Content"

var bucketName, uuidErr = uuid.NewRandom()

func TestMain(m *testing.M) {
	sess := createAwsSession()
	svc := s3.New(sess)

	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName.String()),
	}

	_, err := svc.CreateBucket(createBucketInput)
	if err != nil {
		fmt.Println(err.Error())
	}

	newS3Client(sess, bucketName.String())
	code := m.Run()

	deleteBucketInput := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName.String()),
	}

	_, err = svc.DeleteBucket(deleteBucketInput)
	if err != nil {
		fmt.Println(err.Error())
	}

	os.Exit(code)
}

func TestNewS3Object(t *testing.T) {
	fileName := randomFileName()

	o, err := NewS3Object(fileName)
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

	deleteS3Object(&fileName)
}

func TestS3Object_Exists(t *testing.T) {
	fileName := randomFileName()

	o, err := NewS3Object(fileName)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	err = o.Write([]byte(TestObjectContent))
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if exists := o.Exists(); !exists {
		t.Errorf("Expected TestFile.zip to exist")
	}

	deleteS3Object(&fileName)
}

func TestS3Object_Overridable(t *testing.T) {
	fileName := randomFileName()

	o, err := NewS3Object(fileName)
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

	override := "Some overridden content"
	// override
	err = o.Write([]byte(override))
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	buf, err = o.Read()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if string(buf) != override {
		t.Errorf("Expected object to be '%s', was '%s'", override, string(buf))
		t.Fail()
	}

	deleteS3Object(&fileName)
}

func createAwsSession() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		Config: *aws.NewConfig().WithRegion(endpoints.UsEast1RegionID),
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func randomFileName() string {
	randomUuid, _ := uuid.NewRandom()

	return randomUuid.String() + ".zip"
}

func deleteS3Object(fileName *string) {
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName.String()),
		Key: fileName,
	}
	getS3Client().client.DeleteObject(deleteObjectInput)
}
