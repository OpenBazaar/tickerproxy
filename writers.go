package ticker

import (
	"bytes"
	"io/ioutil"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gocraft/health"
)

// Writer is a callback for data collected from the backend sources
type Writer func(job *health.Job, data []byte) error

// NewFileSystemWriter creates a Writer to writes to a local filesystem
func NewFileSystemWriter(outpath string) Writer {
	return func(job *health.Job, data []byte) error {
		filePath := path.Join(outpath, "rates")
		writerKvs := health.Kvs{"path": filePath}
		err := ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			job.EventErrKv("write.file_system.rates", err, writerKvs)
			return err
		}
		job.EventKv("write.file_system.rates", writerKvs)

		filePath = path.Join(outpath, "whitelist")
		writerKvs = health.Kvs{"path": filePath}
		err = ioutil.WriteFile(filePath, PinnedSymbolsToIDsJSON(), 0644)
		if err != nil {
			job.EventErrKv("write.file_system.whitelist", err, writerKvs)
			return err
		}
		job.EventKv("write.file_system.whitelist", writerKvs)
		return nil
	}
}

// NewS3Writer creates a Writer to writes to AWS S3
func NewS3Writer(region string, bucket string) (Writer, error) {
	creds := credentials.NewEnvCredentials()
	_, err := creds.Get()
	if err != nil {
		return nil, err
	}
	s3CFG := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	s3Client := s3.New(session.New(), s3CFG)

	return func(job *health.Job, data []byte) error {
		_, err := s3Client.PutObject(&s3.PutObjectInput{
			Key:           aws.String("rates"),
			Bucket:        aws.String(bucket),
			Body:          bytes.NewReader(data),
			ContentLength: aws.Int64(int64(len(data))),
			ContentType:   aws.String("application/json"),
		})
		if err != nil {
			job.EventErr("write.s3.rates", err)
			return err
		}
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Key:           aws.String("whitelist"),
			Bucket:        aws.String(bucket),
			Body:          bytes.NewReader(PinnedSymbolsToIDsJSON()),
			ContentLength: aws.Int64(int64(len(PinnedSymbolsToIDsJSON()))),
			ContentType:   aws.String("application/json"),
		})
		if err != nil {
			job.EventErr("write.s3.whitelist", err)
			return err
		}
		job.Event("write.s3")
		return nil
	}, nil
}
