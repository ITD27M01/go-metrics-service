package encryption

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	ErrCouldntReadBody = errors.New("couldn't read body")
)

// BodyDecrypt decrypts request body
func BodyDecrypt(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)

				return
			}

			body, err := io.ReadAll(r.Body)
			switch {
			case errors.Is(err, io.EOF):
				next.ServeHTTP(w, r)

				return
			case err != nil:
				http.Error(w, fmt.Sprintf("Cannot read provided data: %q", err), http.StatusInternalServerError)

				return
			}

			decryptedBody, err := RSADecrypt(body, privateKey)
			if err != nil {
				http.Error(w, fmt.Sprintf("Cannot decrypt provided data: %q", err), http.StatusBadRequest)

				return
			}

			r.Body = ioutil.NopCloser(bytes.NewBuffer(decryptedBody))
			r.ContentLength = int64(len(decryptedBody))
			r.Header.Set("Content-Type", "application/json")

			next.ServeHTTP(w, r)
		})
	}
}

// EncryptRoundTripper encrypt body of request
type EncryptRoundTripper struct {
	proxied   http.RoundTripper
	publicKey *rsa.PublicKey
}

func NewEncryptRoundTripper(proxied http.RoundTripper, publicKey *rsa.PublicKey) *EncryptRoundTripper {
	return &EncryptRoundTripper{
		proxied:   proxied,
		publicKey: publicKey,
	}
}

func (ert *EncryptRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if ert.publicKey == nil || req.Body == nil {
		return ert.proxied.RoundTrip(req)
	}

	body, err := io.ReadAll(req.Body)
	switch {
	case errors.Is(err, io.EOF):
		return ert.proxied.RoundTrip(req)
	case err != nil:
		return nil, ErrCouldntReadBody
	}

	encryptedBody, err := RSAEncrypt(body, ert.publicKey)
	if err != nil {
		return nil, fmt.Errorf("couldn't encrypt body: %w", err)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(encryptedBody))
	req.ContentLength = int64(len(encryptedBody))

	return ert.proxied.RoundTrip(req)
}
