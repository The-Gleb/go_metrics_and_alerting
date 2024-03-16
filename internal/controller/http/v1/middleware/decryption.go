package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type decryptionMiddleware struct {
	privateKey *rsa.PrivateKey
}

func NewDecryptionMiddleware(path string) *decryptionMiddleware {

	privateKeyPEM, err := os.ReadFile("/mnt/d/Programming/Go/src/Study/Practicum/go_metrics_and_alerting/cmd/server/private.pem")
	if err != nil {
		panic(err)
	}
	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		panic(err)
	}

	return &decryptionMiddleware{
		privateKey: privateKey,
	}

}

func (md *decryptionMiddleware) Do(h http.Handler) http.Handler {

	decryptionMiddleware := func(rw http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("decryption middleware working")

		cipher, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		// logger.Log.Debugf("request body is %s", string(cipher))

		// hexDecodedText, err := hex.DecodeString(string(cipher))
		// if err != nil {

		// 	http.Error(rw, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// logger.Log.Debugf("hexDecodedText %s", string(hexDecodedText))

		plainText, err := md.privateKey.Decrypt(rand.Reader, cipher, nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(plainText))

		h.ServeHTTP(rw, r)
	}
	return http.HandlerFunc(decryptionMiddleware)
}
