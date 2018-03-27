package middleware

import (
	"Caw/UserService/utils"
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/anycmon/throttle"
	"net/http"
	"strings"
	"time"
)

const (
	ApplicationJSON = "application/json"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		f = middleware(f)
	}

	return f
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
}

func Logging(log *logrus.Logger) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Infof("--> %s %s", r.Method, r.URL)
			response := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			f(response, r)
			switch response.statusCode {
			case http.StatusOK:
				log.Infof("<-- %d %s", response.statusCode, http.StatusText(response.statusCode))
			default:
				log.Errorf("<-- %d %s", response.statusCode, http.StatusText(response.statusCode))
			}
		}
	}
}

func MustAuth(log *logrus.Logger) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			words := strings.Fields(authorization)
			if len(words) != 2 {
				log.Errorf("Invalid Authorization header %v", authorization)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if words[0] != "Bearer" {
				log.Errorf("Unsupported Authorization type: %v", words[0])
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userClaims, err := utils.DecodeUserToken(words[1])
			if err != nil {
				log.Errorf("Cannot decode user token: %v", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if err = userClaims.Valid(); err != nil {
				log.Errorf("Invalid user token: %v", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "userClaims", userClaims)
			f(w, r.WithContext(ctx))
		}
	}
}

func Produce(supportedAccept string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			accepts := strings.Split(r.Header.Get("Accept"), ",")
			if len(accepts) == 0 || (len(accepts) == 1 && (accepts[0] == "*/*" || accepts[0] == "")) {
				w.Header().Set("Content-Type", supportedAccept)
				f(w, r)
				return
			}
			for _, accept := range accepts {
				if strings.TrimSpace(accept) == supportedAccept {
					w.Header().Set("Content-Type", supportedAccept)
					f(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusNotAcceptable)
		}
	}
}

func Consume(supportedContentType string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			contentTypes := strings.Split(r.Header.Get("Content-Type"), ",")
			for _, contentType := range contentTypes {
				if strings.TrimSpace(contentType) == supportedContentType {
					f(w, r)
					return
				}
			}
			w.WriteHeader(http.StatusUnsupportedMediaType)
		}
	}
}

func CQRS() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			f(w, r)
		}
	}
}

func Throttle(throttle throttle.Throttle) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if throttle.Allow() == false {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			f(w, r)
		}
	}
}

func Per(events int, period time.Duration) float64 {
	return float64(events) / period.Seconds()
}
