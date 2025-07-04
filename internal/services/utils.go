package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sigv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type SigV4RoundTripper struct {
	Transport   http.RoundTripper
	Region      string
	Service     string
	Credentials aws.CredentialsProvider
}

func (s *SigV4RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Always set Accept header to explicitly request JSON
	reqClone.Header.Set("Accept", "application/json")

	// Get AWS credentials
	credentials, err := s.Credentials.Retrieve(req.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS credentials: %w", err)
	}

	// Set session token header if using temporary credentials
	if credentials.SessionToken != "" {
		reqClone.Header.Set("X-Amz-Security-Token", credentials.SessionToken)
	}

	// Handle URL encoding for Prometheus query parameter
	if reqClone.URL.RawQuery != "" {
		q, err := url.ParseQuery(reqClone.URL.RawQuery)
		if err == nil && q.Get("query") != "" {
			originalQuery := q.Get("query")
			encodedQuery := url.QueryEscape(originalQuery)
			q.Set("query", encodedQuery)
			reqClone.URL.RawQuery = q.Encode()
			fmt.Printf("Original query: %s\n", originalQuery)
			fmt.Printf("Encoded query: %s\n", encodedQuery)
			fmt.Printf("Full encoded URL: %s\n", reqClone.URL.String())
		}
	}

	// Prepare the body for signing
	body, err := prepareBody(reqClone)
	if err != nil {
		return nil, err
	}

	// Sign the request using AWS Signature V4
	err = sigv4.NewSigner().SignHTTP(
		reqClone.Context(),
		credentials,
		reqClone,
		hashPayload(body),
		s.Service,
		s.Region,
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Log the full request after signing
	_, err = httputil.DumpRequestOut(reqClone, true)
	if err != nil {
		fmt.Printf("Error dumping request: %v\n", err)
	}

	// Send the request using the underlying transport
	resp, err := s.Transport.RoundTrip(reqClone)
	if err != nil {
		return nil, err
	}

	// For error responses, log the response details
	if resp.StatusCode >= 400 {
		fmt.Printf("Received error response (status %d)\n", resp.StatusCode)
		fmt.Printf("Response headers: %v\n", resp.Header)

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
		} else {
			// Create a new reader with the same content for later consumption
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			fmt.Printf("Error response body:\n%s\n", string(bodyBytes))
		}
	}

	return resp, nil
}

func prepareBody(req *http.Request) (io.ReadSeeker, error) {
	if req.Body == nil {
		return nil, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	err = req.Body.Close()
	if err != nil {
		return nil, err
	}

	body := strings.NewReader(string(bodyBytes))
	req.Body = io.NopCloser(body)
	return body, nil
}

func hashPayload(body io.ReadSeeker) string {
	if body == nil {
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // Empty payload hash
	}

	if _, err := body.Seek(0, io.SeekStart); err != nil {
		return "UNSIGNED-PAYLOAD"
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return "UNSIGNED-PAYLOAD"
	}

	if _, err := body.Seek(0, io.SeekStart); err != nil {
		return "UNSIGNED-PAYLOAD"
	}

	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
