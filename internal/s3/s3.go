package s3

import (
	"bytes"
	"fmt"
	"net/http"
	"pottogether/config"
	"pottogether/pkg/errhandler"
	"pottogether/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

const (
	S3_REGION = "ap-southeast-2"
	S3_BUCKET = "pottogether"
)

var s3Session *session.Session

func InitS3Session() {
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(S3_REGION),
		Credentials: credentials.NewStaticCredentials(
			config.Viper.GetString("AWS_ACCESS_KEY_ID"),
			config.Viper.GetString("AWS_ACCESS_KEY_SECRET"),
			""),
	})
	if err != nil {
		logger.Error("[S3] " + err.Error())
	}
	logger.Info("[S3] Session created")
	s3Session = s
}

func UploadImage(image []byte, filename string) (string, error) {
	filetype := http.DetectContentType(image)
	fileBytes := bytes.NewReader(image)
	fileSize := int64(len(image))
	logger.Info("[S3] Uploading image " + filename)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(S3_BUCKET),
		Key:           aws.String(filename),
		Body:          fileBytes,
		ContentLength: aws.Int64(fileSize),
		ContentType:   aws.String(filetype),
		ACL:           aws.String("public-read"),
	}
	_, err := s3.New(s3Session).PutObject(params)
	if err != nil {
		logger.Error("[S3] " + err.Error())
		return "", err
	}
	logger.Info("[S3] Image uploaded")
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", S3_BUCKET, S3_REGION, filename)
	return url, nil
}

// upload middleware
func UploadMiddleware(c *gin.Context, kind string, filename string) {
	file, err := c.FormFile("image")
	if err != nil {
		errhandler.Error(c, err, "Error retrieving image from the form")
		c.Abort()
		return
	}
	fileBytes, err := file.Open()
	if err != nil {
		errhandler.Error(c, err, "Error opening image")
		c.Abort()
		return
	}
	defer fileBytes.Close()
	buffer := make([]byte, file.Size)
	if _, err := fileBytes.Read(buffer); err != nil {
		errhandler.Error(c, err, "Error reading image")
		c.Abort()
		return
	}
	// upload image to s3
	InitS3Session()
	if filename == "" {
		filename = file.Filename
	}
	path := fmt.Sprintf("%s/%s", kind, filename)
	url, err := UploadImage(buffer, path)
	if err != nil {
		errhandler.Error(c, err, "Error uploading image to s3")
		c.Abort()
		return
	}
	c.Set("image", url)
}
