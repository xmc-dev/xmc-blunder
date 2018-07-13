package s3

import (
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/xmc-core/db/models/attachment"
	"github.com/xmc-dev/xmc/xmc-core/service"
	"gopkg.in/h2non/filetype.v1"
)

var client *minio.Client

var log = logrus.WithField("prefix", "s3")
var bucketName string

// Init initializes the S3 connection
func Init() error {
	var err error

	endpoint := service.MainService.S3Endpoint
	accKey := service.MainService.S3AccessKeyID
	secret := service.MainService.S3SecretAccessKey
	ssl := service.MainService.S3UseSSL

	client, err = minio.New(endpoint, accKey, secret, ssl)
	if err != nil {
		return errors.Wrap(err, "couldn't create S3 object")
	}
	log.WithFields(logrus.Fields{
		"endpoint": endpoint,
		"accKey":   accKey,
		"ssl":      ssl,
	}).Info("Connected to s3")

	bucketName = service.MainService.S3BucketName
	loc := service.MainService.S3BucketLocation
	err = client.MakeBucket(bucketName, loc)
	if err != nil {
		ok, err2 := client.BucketExists(bucketName)
		if err2 != nil {
			return errors.Wrap(err2, "couldn't check if bucket exists")
		} else if err2 == nil && !ok {
			return errors.Wrap(err, "couldn't make bucket")
		}
	}
	log.WithFields(logrus.Fields{
		"bucket":   bucketName,
		"location": loc,
	}).Info("We have a bucket")

	return nil
}

func s3Name(att *attachment.Attachment) string {
	return fmt.Sprintf("%s/%s_%s", att.ObjectID, att.ID, att.Filename)
}

// UploadAttachment uploads the file to S3
func UploadAttachment(att *attachment.Attachment, contents io.Reader, length int64) (string, error) {
	objName := s3Name(att)
	mime, err := filetype.MatchReader(contents)
	kind := "application/octet-stream"
	if err == nil && len(mime.MIME.Value) > 0 {
		kind = mime.MIME.Value
	}
	log.WithFields(logrus.Fields{
		"mime":    kind,
		"objName": objName,
	}).Infof("Going to upload to S3")
	_, err = client.PutObject(bucketName, objName, contents, length, minio.PutObjectOptions{
		ContentType: kind,
	})
	if err != nil {
		return "", errors.Wrap(err, "couldn't put S3 object")
	}

	log.WithFields(logrus.Fields{
		"objName": objName,
		"kind":    kind,
	}).Info("Successfully uploaded attachment to S3")

	return objName, nil
}

// GetURL returns the URL of the attachment file
func GetURL(att *attachment.Attachment, contentDisposition string) (*url.URL, error) {
	objName := s3Name(att)
	if len(contentDisposition) == 0 {
		contentDisposition = att.Filename
	}
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\""+contentDisposition+"\"")

	u, err := client.PresignedGetObject(bucketName, objName, 15*time.Minute, reqParams)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get presigned object")
	}
	return u, nil
}

// DeleteAttachment deletes the file from S3
func DeleteAttachment(att *attachment.Attachment) error {
	objName := s3Name(att)
	err := client.RemoveObject(bucketName, objName)

	if err != nil {
		return errors.Wrap(err, "couldn't remove S3 object")
	}

	log.WithFields(logrus.Fields{"objName": objName}).Info("Successfully deleted attachment from S3")

	return nil
}

func RenameAttachment(att *attachment.Attachment, newFilename string) (string, error) {
	a := &attachment.Attachment{}
	*a = *att
	a.Filename = newFilename

	objName := s3Name(att)
	objNewName := s3Name(a)

	src := minio.NewSourceInfo(bucketName, objName, nil)
	dst, err := minio.NewDestinationInfo(bucketName, objNewName, nil, nil)
	if err != nil {
		return "", errors.Wrap(err, "couldn't get destination info to initiate rename")
	}

	err = client.CopyObject(dst, src)
	if err != nil {
		return "", errors.Wrap(err, "couldn't copy object for rename")
	}

	err = client.RemoveObject(bucketName, objName)
	if err != nil {
		return "", errors.Wrap(err, "couldn't remove object for rename")
	}

	log.WithFields(logrus.Fields{
		"objName":    objName,
		"objNewName": objNewName,
	}).Info("Successfully renamed attachment in S3")

	return objNewName, nil
}

// Deinit deinitializes internal stuff
func Deinit() {
}
