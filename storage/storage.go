// Package storage provides direct S3 / MinIO / R2 cloud storage upload capabilities for packaged media streams.
package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// S3Config contains the connection and authentication details for S3-compatible cloud storage.
type S3Config struct {
	Endpoint        string
	Bucket          string
	Region          string
	AccessKey       string
	SecretKey       string
	KeyPrefix       string
	ConcurrentFiles int
	UseSSL          bool
}

// Normalize sets default values for S3Config.
func (c *S3Config) Normalize() {
	if c.Region == "" {
		c.Region = "us-east-1"
	}
	if c.ConcurrentFiles <= 0 {
		c.ConcurrentFiles = 5
	}
}

// ContentTypeForFile returns the optimal MIME Content-Type header for streaming files.
func ContentTypeForFile(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".mpd":
		return "application/dash+xml"
	case ".m4s":
		return "video/iso.segment"
	case ".ts":
		return "video/mp2t"
	case ".vtt":
		return "text/vtt"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".key":
		return "application/octet-stream"
	default:
		return "application/octet-stream"
	}
}

// CacheControlForFile returns recommended HTTP cache control headers.
func CacheControlForFile(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".m3u8", ".mpd":
		return "no-cache, no-store, must-revalidate"
	case ".m4s", ".ts", ".jpg", ".png":
		return "public, max-age=31536000, immutable"
	default:
		return "public, max-age=86400"
	}
}

// Signer generates AWS Signature Version 4 authorization headers.
type Signer struct {
	Region    string
	AccessKey string
	SecretKey string
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// Sign calculates the AWS SigV4 Authorization header for an HTTP PUT request.
func (s *Signer) Sign(req *http.Request, payloadHash string, t time.Time) {
	dateStamp := t.UTC().Format("20060102")
	amzDate := t.UTC().Format("20060102T150405Z")

	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	canonicalURI := req.URL.EscapedPath()
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", req.Host, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		"", // query string
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	canonicalRequestHash := sha256Hex([]byte(canonicalRequest))

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, s.Region)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		canonicalRequestHash,
	}, "\n")

	// Calculate signature
	kDate := hmacSHA256([]byte("AWS4"+s.SecretKey), dateStamp)
	kRegion := hmacSHA256(kDate, s.Region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.AccessKey, credentialScope, signedHeaders, signature)

	req.Header.Set("Authorization", authHeader)
}

// UploadDirectory walks the directory and uploads all stream assets to the configured S3 bucket.
func UploadDirectory(ctx context.Context, dir string, cfg S3Config, client *http.Client) error {
	cfg.Normalize()
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk directory: %w", err)
	}

	var endpoint string
	if cfg.Endpoint != "" {
		endpoint = strings.TrimSuffix(cfg.Endpoint, "/")
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			scheme := "https://"
			if !cfg.UseSSL {
				scheme = "http://"
			}
			endpoint = scheme + endpoint
		}
	} else {
		endpoint = fmt.Sprintf("https://s3.%s.amazonaws.com", cfg.Region)
	}

	signer := &Signer{
		Region:    cfg.Region,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}

	fileChan := make(chan string, len(files))
	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	for i := 0; i < cfg.ConcurrentFiles; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				relPath, err := filepath.Rel(dir, filePath)
				if err != nil {
					errChan <- err
					return
				}

				key := relPath
				if cfg.KeyPrefix != "" {
					key = strings.TrimSuffix(cfg.KeyPrefix, "/") + "/" + relPath
				}

				data, err := os.ReadFile(filePath)
				if err != nil {
					errChan <- fmt.Errorf("read file %s: %w", filePath, err)
					return
				}

				uploadURL := fmt.Sprintf("%s/%s/%s", endpoint, cfg.Bucket, url.PathEscape(key))
				req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(data))
				if err != nil {
					errChan <- fmt.Errorf("create upload request for %s: %w", key, err)
					return
				}

				req.Header.Set("Content-Type", ContentTypeForFile(filePath))
				req.Header.Set("Cache-Control", CacheControlForFile(filePath))

				payloadHash := sha256Hex(data)
				signer.Sign(req, payloadHash, time.Now())

				resp, err := client.Do(req)
				if err != nil {
					errChan <- fmt.Errorf("upload %s failed: %w", key, err)
					return
				}
				body, _ := io.ReadAll(resp.Body)
				_ = resp.Body.Close()

				if resp.StatusCode >= 300 {
					errChan <- fmt.Errorf("upload %s returned status %d: %s", key, resp.StatusCode, string(body))
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}
