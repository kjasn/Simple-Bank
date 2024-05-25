package gapi

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(ctx context.Context, req any, 
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (resp any, err error) {
	startTime := time.Now()
	resp, err = handler(ctx, req)
	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}
	duration := time.Since(startTime)

	logger := log.Info()
	if err != nil {
		logger.Err(err)
	}

	logger.
	Str("Protocol", "gRPC").
	Int("Status_code", int(statusCode)).
	Str("Status_text", statusCode.String()).
	Str("Method", info.FullMethod).
	Dur("Duration", duration).
	Msg("Received a gRPC request")
	return resp, err
}


type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body []byte
	StatusText string
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.StatusText = http.StatusText(statusCode)
	rec.ResponseWriter.WriteHeader(statusCode)
}

func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HttpLogger (handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		startTime := time.Now()

		rec := &ResponseRecorder{
			ResponseWriter: rsp,
			StatusCode: http.StatusOK,
			StatusText: "OK",
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()
		if rec.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rec.Body)
		}

		logger.
		Str("Protocol", "HTTP").
		Str("Method", req.Method).
		Str("Path", req.RequestURI).
		Int("StatusCode", rec.StatusCode).
		// Str("StatusText", http.StatusText(rec.StatusCode)).
		Str("StatusText", rec.StatusText).
		Dur("Duration", duration).
		Msg("Received a HTTP request")
	})
}