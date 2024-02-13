package grpc

import (
	"context"
	"fmt"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/cookie"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/models"
	"github.com/JustWorking42/shortener-go-yandex/internal/app/storage"
	"github.com/JustWorking42/shortener-go-yandex/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerService struct {
	proto.UnimplementedShortenerServiceServer
	app *app.App
}

func NewShortenerService(app *app.App) *ShortenerService {
	return &ShortenerService{app: app}
}

var _ proto.ShortenerServiceServer = &ShortenerService{}

func (s *ShortenerService) ShortUrl(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	userID := ctx.Value(cookie.UserID("UserID")).(string)
	savedURL, err := s.app.Repository.SaveURL(ctx, req.Url, userID)
	if err != nil {
		s.app.Logger.Sugar().Error(err)
		if savedURL != (storage.SavedURL{}) {
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("%s/%s", s.app.RedirectHost, savedURL.ShortURL))
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &proto.ShortenURLResponse{ShortUrl: fmt.Sprintf("%s/%s", s.app.RedirectHost, savedURL.ShortURL)}, nil
}

func (s *ShortenerService) ShortUrlsBatch(ctx context.Context, req *proto.ShortenURLsBatchRequest) (*proto.ShortenURLsBatchResponse, error) {
	userID := ctx.Value(cookie.UserID("UserID")).(string)
	var originalURLsSlice []models.RequestShortenerURLBatch

	for _, url := range req.Urls {
		originalURLsSlice = append(originalURLsSlice, models.RequestShortenerURLBatch{
			ID:  url.Id,
			URL: url.Url,
		})
	}

	savedURLsSlice, err := s.app.Repository.SaveURLArray(ctx, originalURLsSlice, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var shortURLs []*proto.ResponseShortenerURLBatch
	for _, savedURL := range savedURLsSlice {
		shortURLs = append(shortURLs, &proto.ResponseShortenerURLBatch{
			Id:       savedURL.ID,
			ShortUrl: fmt.Sprintf("%s/%s", s.app.RedirectHost, savedURL.URL),
		})
	}

	return &proto.ShortenURLsBatchResponse{Urls: shortURLs}, nil
}

func (s *ShortenerService) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	savedURL, err := s.app.Repository.GetURL(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "URL not found")
	}
	if savedURL.IsDeleted {
		return nil, status.Error(codes.Unavailable, "URL has been deleted")
	}
	return &proto.GetURLResponse{OriginalUrl: savedURL.OriginalURL}, nil
}

func (s *ShortenerService) GetUserURLs(ctx context.Context, req *emptypb.Empty) (*proto.GetUserURLsResponse, error) {
	userID := ctx.Value(cookie.UserID("UserID")).(string)
	urls, err := s.app.Repository.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var userURLs []*proto.GetUserURLs
	for _, url := range urls {
		userURLs = append(userURLs, &proto.GetUserURLs{
			ShortUrl:    fmt.Sprintf("%s/%s", s.app.RedirectHost, url.ShortURL),
			OriginalUrl: url.OriginalURL,
		})
	}

	return &proto.GetUserURLsResponse{Urls: userURLs}, nil
}

func (s *ShortenerService) DeleteURLs(ctx context.Context, req *proto.DeleteURLsRequest) (*emptypb.Empty, error) {
	userID := ctx.Value(cookie.UserID("UserID")).(string)
	if err := s.app.Repository.DeleteURLs(ctx, userID, req.Urls); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *ShortenerService) GetStats(ctx context.Context, req *emptypb.Empty) (*proto.GetStatsResponse, error) {
	stats, err := s.app.Repository.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.GetStatsResponse{
		UrlsCount:  int32(stats.URLs),
		UsersCount: int32(stats.Users),
	}, nil
}

func (s *ShortenerService) PingDB(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.app.Repository.PingDB(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "database ping failed")
	}
	return &emptypb.Empty{}, nil
}
