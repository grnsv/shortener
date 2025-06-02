// Package pb provides the gRPC server implementation for the URL shortener service.
package pb

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
)

// GRPCShortenerServer implements the gRPC Shortener service.
type GRPCShortenerServer struct {
	UnimplementedShortenerServer
	shortener service.Shortener // Service for URL shortening logic
	logger    logger.Logger     // Logger for error and info messages
}

// NewGRPCShortenerServer creates a new instance of GRPCShortenerServer.
func NewGRPCShortenerServer(shortener service.Shortener, logger logger.Logger) *GRPCShortenerServer {
	return &GRPCShortenerServer{shortener: shortener, logger: logger}
}

// ShortenURL shortens a given URL for the authenticated user.
func (s *GRPCShortenerServer) ShortenURL(ctx context.Context, in *ShortenRequest) (*ShortenResponse, error) {
	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok {
		s.logger.Error("user ID not found in context")
		return nil, status.Error(codes.Unauthenticated, "Empty userID")
	}

	if in.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "Empty url")
	}

	shortURL, alreadyExists, err := s.shortener.ShortenURL(ctx, in.Url, userID)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	if alreadyExists {
		return nil, status.Error(codes.AlreadyExists, codes.AlreadyExists.String())
	}

	return &ShortenResponse{Result: shortURL}, nil
}

// ShortenBatch shortens multiple URLs in a batch for the authenticated user.
func (s *GRPCShortenerServer) ShortenBatch(ctx context.Context, in *BatchRequest) (*BatchResponse, error) {
	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok {
		s.logger.Error("user ID not found in context")
		return nil, status.Error(codes.Unauthenticated, "Empty userID")
	}

	if in == nil || len(in.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Empty batch request")
	}

	req := make([]models.BatchRequestItem, len(in.Items))
	for i, item := range in.Items {
		req[i] = models.BatchRequestItem{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		}
	}

	resp, err := s.shortener.ShortenBatch(ctx, req, userID)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := BatchResponse{
		Items: make([]*BatchResponseItem, len(resp)),
	}
	for i, item := range resp {
		out.Items[i] = &BatchResponseItem{
			CorrelationId: item.CorrelationID,
			ShortUrl:      item.ShortURL,
		}
	}

	return &out, nil
}

// ExpandURL expands a shortened URL ID to its original URL.
func (s *GRPCShortenerServer) ExpandURL(ctx context.Context, in *ExpandRequest) (*ExpandResponse, error) {
	if in == nil || in.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Empty id")
	}

	url, err := s.shortener.ExpandURL(ctx, in.Id)
	if err != nil {
		if errors.Is(err, storage.ErrDeleted) {
			return nil, status.Error(codes.NotFound, "URL deleted")
		}

		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ExpandResponse{Url: url}, nil
}

// PingDB checks the health of the database/storage.
func (s *GRPCShortenerServer) PingDB(ctx context.Context, in *Empty) (*Empty, error) {
	if err := s.shortener.PingStorage(ctx); err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &Empty{}, nil
}

// GetURLs retrieves all shortened URLs for the authenticated user.
func (s *GRPCShortenerServer) GetURLs(ctx context.Context, in *Empty) (*GetURLsResponse, error) {
	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok {
		s.logger.Error("user ID not found in context")
		return nil, status.Error(codes.Unauthenticated, "Empty userID")
	}

	urls, err := s.shortener.GetAll(ctx, userID)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := make([]*URLItem, len(urls))
	for i, u := range urls {
		resp[i] = &URLItem{
			UserId:      u.UserID,
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		}
	}

	return &GetURLsResponse{Urls: resp}, nil
}

// DeleteURLs deletes multiple shortened URLs for the authenticated user.
func (s *GRPCShortenerServer) DeleteURLs(ctx context.Context, in *DeleteURLsRequest) (*Empty, error) {
	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)
	if !ok {
		s.logger.Error("user ID not found in context")
		return nil, status.Error(codes.Unauthenticated, "Empty userID")
	}

	if in == nil || len(in.ShortUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Empty short_urls")
	}

	err := s.shortener.DeleteMany(ctx, userID, in.ShortUrls)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &Empty{}, nil
}

// GetStats returns statistics about the number of URLs and users.
func (s *GRPCShortenerServer) GetStats(ctx context.Context, in *Empty) (*StatsResponse, error) {
	stats, err := s.shortener.GetStats(ctx)
	if err != nil {
		s.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &StatsResponse{Urls: int32(stats.URLsCount), Users: int32(stats.UsersCount)}, nil
}
