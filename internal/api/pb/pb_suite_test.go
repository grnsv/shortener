package pb_test

import (
	"context"
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/golang/mock/gomock"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/api/pb"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/mocks"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/storage"
)

func TestPb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pb Suite")
}

var _ = Describe("GRPCShortenerServer", func() {
	var (
		ctrl          *gomock.Controller
		mockShortener *mocks.MockShortener
		log           logger.Logger
		server        *grpc.Server
		client        pb.ShortenerClient
		conn          *grpc.ClientConn
		listener      net.Listener
		err           error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockShortener = mocks.NewMockShortener(ctrl)
		log, err = logger.New("testing")
		Expect(err).To(BeNil())
		listener, err = net.Listen("tcp", ":0")
		Expect(err).To(BeNil())
		server = grpc.NewServer(grpc.UnaryInterceptor(middleware.GRPCAuthenticateInterceptor("secret", log)))
		pb.RegisterShortenerServer(server, pb.NewGRPCShortenerServer(mockShortener, log))
		go func() {
			serverErr := server.Serve(listener)
			Expect(serverErr).To(BeNil())
		}()
		conn, err = grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).ToNot(HaveOccurred())
		client = pb.NewShortenerClient(conn)
	})

	AfterEach(func() {
		server.Stop()
		err = conn.Close()
		Expect(err).To(BeNil())
		ctrl.Finish()
	})

	Context("PingDB", func() {
		When("the database is reachable", func() {
			It("returns OK", func() {
				mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)
				_, err := client.PingDB(context.Background(), &pb.Empty{})
				Expect(err).To(BeNil())
			})
		})
		When("the database is unreachable", func() {
			It("returns an error", func() {
				mockShortener.EXPECT().PingStorage(gomock.Any()).Return(assert.AnError)
				_, err := client.PingDB(context.Background(), &pb.Empty{})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Authenticate", func() {
		When("request does not have token", func() {
			It("returns new token in metadata", func() {
				mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)
				var header metadata.MD

				_, err := client.PingDB(context.Background(), &pb.Empty{}, grpc.Header(&header))
				Expect(err).To(BeNil())
				token := header.Get("token")
				Expect(token).ToNot(BeEmpty())
				Expect(token[0]).To(HavePrefix("ey"))
			})
		})
		When("request has invalid token", func() {
			It("returns new token in metadata", func() {
				mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)
				ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", "invalid"))
				var header metadata.MD

				_, err := client.PingDB(ctx, &pb.Empty{}, grpc.Header(&header))
				Expect(err).To(BeNil())
				token := header.Get("token")
				Expect(token).ToNot(BeEmpty())
				Expect(token[0]).To(HavePrefix("ey"))
			})
		})
		When("request has empty token", func() {
			It("returns new token in metadata", func() {
				mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)
				ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", ""))
				var header metadata.MD

				_, err := client.PingDB(ctx, &pb.Empty{}, grpc.Header(&header))
				Expect(err).To(BeNil())
				token := header.Get("token")
				Expect(token).ToNot(BeEmpty())
				Expect(token[0]).To(HavePrefix("ey"))
			})
		})
		When("request has empty user ID", func() {
			It("returns status Unauthenticated", func() {
				jwtString, err := middleware.BuildJWTString("secret", "")
				Expect(err).To(BeNil())
				ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", jwtString))
				var header metadata.MD

				_, err = client.PingDB(ctx, &pb.Empty{}, grpc.Header(&header))
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.Unauthenticated))
				Expect(header).To(BeNil())
			})
		})
	})

	Context("ShortenURL", func() {
		const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		var ctx context.Context
		BeforeEach(func() {
			jwtString, err := middleware.BuildJWTString("secret", userID)
			Expect(err).To(BeNil())
			ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", jwtString))
		})
		It("returns short url", func() {
			mockShortener.EXPECT().ShortenURL(gomock.Any(), "http://example.com", userID).Return("short", false, nil)
			resp, err := client.ShortenURL(ctx, &pb.ShortenRequest{Url: "http://example.com"})
			Expect(err).To(BeNil())
			Expect(resp.Result).To(Equal("short"))
		})
		When("url exists", func() {
			It("returns AlreadyExists", func() {
				mockShortener.EXPECT().ShortenURL(gomock.Any(), "http://example.com", userID).Return("short", true, nil)
				_, err := client.ShortenURL(ctx, &pb.ShortenRequest{Url: "http://example.com"})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.AlreadyExists))
			})
		})
		When("url is empty", func() {
			It("returns error", func() {
				_, err := client.ShortenURL(ctx, &pb.ShortenRequest{Url: ""})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
			})
		})
	})

	Context("GetURLs", func() {
		const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		var ctx context.Context
		BeforeEach(func() {
			jwtString, err := middleware.BuildJWTString("secret", userID)
			Expect(err).To(BeNil())
			ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", jwtString))
		})
		When("storage does not have stored urls", func() {
			It("returns empty list", func() {
				mockShortener.EXPECT().GetAll(gomock.Any(), userID).Return(nil, nil)
				resp, err := client.GetURLs(ctx, &pb.Empty{})
				Expect(err).To(BeNil())
				Expect(resp.Urls).To(BeEmpty())
			})
		})
		When("storage has stored urls", func() {
			It("returns OK", func() {
				urls := []models.URL{
					{
						UUID:        "00000000-0000-0000-0000-000000000001",
						ShortURL:    "00000001",
						OriginalURL: "http://example.com/1",
					},
					{
						UUID:        "00000000-0000-0000-0000-000000000002",
						ShortURL:    "00000002",
						OriginalURL: "http://example.com/2",
					},
				}
				mockShortener.EXPECT().GetAll(gomock.Any(), userID).Return(urls, nil)
				resp, err := client.GetURLs(ctx, &pb.Empty{})
				Expect(err).To(BeNil())
				Expect(len(urls)).To(Equal(len(resp.Urls)))
			})
		})
	})

	Context("ShortenBatch", func() {
		const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		var ctx context.Context
		BeforeEach(func() {
			jwtString, err := middleware.BuildJWTString("secret", userID)
			Expect(err).To(BeNil())
			ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", jwtString))
		})
		When("batch request is valid", func() {
			It("returns batch response", func() {
				modelsReq := models.BatchRequest{
					{CorrelationID: "1", OriginalURL: "http://example.com/1"},
					{CorrelationID: "2", OriginalURL: "http://example.com/2"},
				}
				modelsResp := models.BatchResponse{
					{CorrelationID: "1", ShortURL: "http://localhost:8080/short1"},
					{CorrelationID: "2", ShortURL: "http://localhost:8080/short2"},
				}
				mockShortener.EXPECT().ShortenBatch(gomock.Any(), modelsReq, userID).Return(modelsResp, nil)

				resp, err := client.ShortenBatch(ctx, &pb.BatchRequest{Items: []*pb.BatchRequestItem{
					{CorrelationId: modelsReq[0].CorrelationID, OriginalUrl: modelsReq[0].OriginalURL},
					{CorrelationId: modelsReq[1].CorrelationID, OriginalUrl: modelsReq[1].OriginalURL},
				}})
				Expect(err).To(BeNil())
				Expect(resp.Items).To(HaveLen(2))
				Expect(resp.Items[0].CorrelationId).To(Equal(modelsResp[0].CorrelationID))
				Expect(resp.Items[0].ShortUrl).To(Equal(modelsResp[0].ShortURL))
				Expect(resp.Items[1].CorrelationId).To(Equal(modelsResp[1].CorrelationID))
				Expect(resp.Items[1].ShortUrl).To(Equal(modelsResp[1].ShortURL))
			})
		})
		When("batch request is empty", func() {
			It("returns error", func() {
				_, err := client.ShortenBatch(ctx, &pb.BatchRequest{Items: nil})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
			})
		})
	})

	Context("DeleteURLs", func() {
		const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		var ctx context.Context
		BeforeEach(func() {
			jwtString, err := middleware.BuildJWTString("secret", userID)
			Expect(err).To(BeNil())
			ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", jwtString))
		})
		When("delete request is valid", func() {
			It("returns OK", func() {
				shortURLs := []string{"short1", "short2"}
				mockShortener.EXPECT().DeleteMany(gomock.Any(), userID, shortURLs).Return(nil)
				_, err := client.DeleteURLs(ctx, &pb.DeleteURLsRequest{ShortUrls: shortURLs})
				Expect(err).To(BeNil())
			})
		})
		When("delete request is empty", func() {
			It("returns error", func() {
				_, err := client.DeleteURLs(ctx, &pb.DeleteURLsRequest{ShortUrls: nil})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
			})
		})
		When("delete returns error", func() {
			It("returns error", func() {
				shortURLs := []string{"short1"}
				mockShortener.EXPECT().DeleteMany(gomock.Any(), userID, shortURLs).Return(assert.AnError)
				_, err := client.DeleteURLs(ctx, &pb.DeleteURLsRequest{ShortUrls: shortURLs})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.Internal))
			})
		})
	})

	Context("GetStats", func() {
		const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		var ctx context.Context
		BeforeEach(func() {
			jwtString, err := middleware.BuildJWTString("secret", userID)
			Expect(err).To(BeNil())
			ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", jwtString))
		})
		When("stats are available", func() {
			It("returns stats", func() {
				stats := &models.Stats{URLsCount: 42, UsersCount: 7}
				mockShortener.EXPECT().GetStats(gomock.Any()).Return(stats, nil)
				resp, err := client.GetStats(ctx, &pb.Empty{})
				Expect(err).To(BeNil())
				Expect(resp.Urls).To(Equal(int32(stats.URLsCount)))
				Expect(resp.Users).To(Equal(int32(stats.UsersCount)))
			})
		})
		When("service returns error", func() {
			It("returns error", func() {
				mockShortener.EXPECT().GetStats(gomock.Any()).Return(nil, assert.AnError)
				_, err := client.GetStats(ctx, &pb.Empty{})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.Internal))
			})
		})
	})

	Context("ExpandURL", func() {
		const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		var ctx context.Context
		BeforeEach(func() {
			jwtString, err := middleware.BuildJWTString("secret", userID)
			Expect(err).To(BeNil())
			ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("token", jwtString))
		})
		When("short URL exists", func() {
			It("returns the original URL", func() {
				mockShortener.EXPECT().ExpandURL(gomock.Any(), "short1").Return("http://example.com/1", nil)
				resp, err := client.ExpandURL(ctx, &pb.ExpandRequest{Id: "short1"})
				Expect(err).To(BeNil())
				Expect(resp.Url).To(Equal("http://example.com/1"))
			})
		})
		When("short URL does not exist", func() {
			It("returns NotFound", func() {
				mockShortener.EXPECT().ExpandURL(gomock.Any(), "deleted").Return("", storage.ErrDeleted)
				_, err := client.ExpandURL(ctx, &pb.ExpandRequest{Id: "deleted"})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.NotFound))
			})
		})
		When("short URL is empty", func() {
			It("returns InvalidArgument", func() {
				_, err := client.ExpandURL(ctx, &pb.ExpandRequest{Id: ""})
				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
			})
		})
	})
})
