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

func CheckSignature(signKey []byte, handleFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(signKey) == 0 {
			handleFunc(w, r)
			return
		}
		logger.Log.Debug(string(signKey))

		gotSign, err := hex.DecodeString(r.Header.Get("HashSHA256"))
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

		if !hmac.Equal(sign, gotSign) {
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
		handleFunc(&srw, r)

	}
}
