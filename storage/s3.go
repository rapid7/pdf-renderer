package storage

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rapid7/pdf-renderer/cfg"
)

type s3Client struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
}

// Variables needed to enforce singleton pattern for S3Client
var (
	client *s3Client
)

// singleton client
func getS3Client() (*s3Client) {
	if client == nil {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		newS3Client(sess, cfg.Config().S3Bucket())
	}

	return client
}

func newS3Client(sess *session.Session, bucket string) {
	svc := s3.New(sess)

	u := s3manager.NewUploader(sess)
	d := s3manager.NewDownloader(sess)

	client = &s3Client{
		client:     svc,
		uploader:   u,
		downloader: d,
		bucket:     bucket,
	}
}

type s3Object struct {
	fileName string
	s3client *s3Client
}

func NewS3Object(fileName string) (*s3Object, error) {
	client := getS3Client()

	return &s3Object{
		fileName: fileName,
		s3client: client,
	}, nil

}

func (o *s3Object) FileName() string {
	return o.fileName
}

func (o *s3Object) Write(data []byte) error {
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(o.s3client.bucket),
		Key:    aws.String(o.FileName()),
		Body:   bytes.NewReader(data),
	}

	// Perform the upload with supplied params.
	_, err := o.s3client.uploader.Upload(upParams)
	if err != nil {
		return err
	}

	return nil
}

func (o *s3Object) Read() ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})

	downParams := &s3.GetObjectInput{
		Bucket: aws.String(o.s3client.bucket),
		Key:    aws.String(o.FileName()),
	}

	_, err := o.s3client.downloader.Download(buf, downParams)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

func (o *s3Object) Exists() bool {
	input := &s3.GetObjectInput{
		Bucket: aws.String(o.s3client.bucket),
		Key:    aws.String(o.FileName()),
	}

	_, err := o.s3client.client.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				log.Errorf(fmt.Sprintf("%v error: the key '%v' does not exist", s3.ErrCodeNoSuchKey, o.FileName()))
			default:
				log.Error(aerr.Error())
			}
		} else {
			log.Error(err.Error())
		}
		return false
	}
	return true
}
