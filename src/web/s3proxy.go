package web

import (
	"net/url"
	"strings"
	"time"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
)

var awsSess = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))

func S3Proxy(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		hostStr := c.Request().Host
		if strings.HasPrefix(hostStr, "static.") {
			return S3Handle(c)
		}
		return next(c)
	}
}

func S3Handle(c echo.Context) error {
	file := c.Request().URL.Path

	s3Req, _ := s3.New(awsSess).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("impactclient-builds"),
		Key:    aws.String(file),
	})

	s3PresignedURL, err := s3Req.Presign(1 * time.Minute)
	if err != nil {
		return err
	}

	target, err := url.Parse(s3PresignedURL)
	if err != nil {
		return err
	}

	// https://stackoverflow.com/questions/7071763/max-value-for-cache-control-header-in-http
	c.Response().Header().Set("Cache-Control", "max-age=31536000")

	doProxy(c, target)
	return nil
}
