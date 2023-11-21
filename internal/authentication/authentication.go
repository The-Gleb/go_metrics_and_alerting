package authentication

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type signingResponseWriter struct {
	http.ResponseWriter
	key []byte
}

func (w *signingResponseWriter) Write(b []byte) (int, error) {
	h := hmac.New(sha256.New, w.key)
	_, err := h.Write(b)
	if err != nil {
		http.Error(w.ResponseWriter, err.Error(), http.StatusInternalServerError)
	}
	sign := h.Sum(nil)
	encodedSign := hex.EncodeToString(sign)
	w.ResponseWriter.Header().Set("HashSHA256", encodedSign)
	n, err := w.ResponseWriter.Write(b)
	return n, err
}

func CheckSignature(signKey []byte, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if len(signKey) == 0 || r.Header.Get("Hash") == "none" || r.Header.Get("Hashsha256") == "" {
			logger.Log.Debug("key is empty")
			next.ServeHTTP(w, r)
			return
		}

		if r.Header.Get("Hash") == "none" {
			logger.Log.Debug("key is empty")
			next.ServeHTTP(w, r)
			return
		}

		gotSign, err := hex.DecodeString(r.Header.Get("HashSHA256"))
		// gotSign := r.Header.Get("HashSHA256")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		h := hmac.New(sha256.New, signKey)

		data, _ := io.ReadAll(r.Body)
		r.Body.Close()
		_, err = h.Write(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		sign := h.Sum(nil)

		logger.Log.Debug("sign key is ", string(signKey))
		logger.Log.Debug("received body is ", string(data))
		logger.Log.Debug("received hex signature: ", r.Header.Get("HashSHA256"))
		logger.Log.Debug("calculated hex signature: ", hex.EncodeToString(sign))
		logger.Log.Debug("Headers: ", r.Header)

		if !hmac.Equal(sign, []byte(gotSign)) {
			logger.Log.Debug("hash signatures are not equal")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(data))
		srw := signingResponseWriter{
			ResponseWriter: w,
			key:            signKey,
		}
		next.ServeHTTP(&srw, r)

	}
	return http.HandlerFunc(fn)
}
