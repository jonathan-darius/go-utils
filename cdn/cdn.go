package cdn

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// Config cdn
type Config struct {
	Key     []byte
	SaltKey []byte
	Host    string
}

// Image definition
type Image struct {
	Url       string
	Resize    string
	Width     int
	Height    int
	X         int
	Y         int
	Gravity   string
	Enlarge   int
	Extension string
}

type S3 struct {
	BucketName string
	Path       string
}

func New(host, key, salt string) (*Config, error) {
	var keyBin, saltBin []byte
	var err error

	if keyBin, err = hex.DecodeString(key); err != nil {
		return nil, errors.New("Key expected to be hex-encoded string")
	}

	if saltBin, err = hex.DecodeString(salt); err != nil {
		return nil, errors.New("Salt expected to be hex-encoded string")
	}

	return &Config{
		Key:     keyBin,
		SaltKey: saltBin,
		Host:    host,
	}, nil

}

func (c *Config) GetUrl(img *Image) string {
	// key := ""
	// salt := ""

	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(img.Url))

	resize := "fill"
	width := 300
	height := 0
	gravity := "no"
	enlarge := 1
	extension := "webp"

	if img.Resize != "" {
		resize = img.Resize
	}

	if img.Width != 0 {
		width = img.Width
	}
	if img.Height != 0 {
		height = img.Height
	}
	if img.Gravity != "" {
		gravity = img.Gravity
	}
	if img.Enlarge != 00 {
		enlarge = img.Enlarge
	}
	if img.Extension != "" {
		extension = img.Extension
	}

	path := fmt.Sprintf("/%s/%d/%d/%s/%d/%s.%s", resize, width, height, gravity, enlarge, encodedURL, extension)

	mac := hmac.New(sha256.New, c.Key)
	mac.Write(c.SaltKey)
	mac.Write([]byte(path))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s/%s%s \n", c.Host, signature, path)
}

func (c *Config) GetS3Url(s3 *S3) string {
	return fmt.Sprintf("s3://%s/%s", s3.BucketName, s3.Path)
}
