// Package grpc provides gRPC handlers for the URL shortening service.
package grpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/service"
	grpcshortener "github.com/bezjen/shortener/pkg/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Handler implements gRPC service for URL shortening operations.
// It provides methods for shortening URLs, expanding short URLs, and listing user URLs.
type Handler struct {
	grpcshortener.UnimplementedShortenerServiceServer

	cfg          config.Config
	logger       *logger.Logger
	authorizer   service.Authorizer
	shortener    service.Shortener
	auditService service.AuditService
}

// NewHandler creates a new instance of Handler.
//
// Parameters:
//   - cfg: Application configuration settings
//   - logger: Logger instance for application logging
//   - authorizer: Service for user authentication and authorization
//   - shortener: URL shortening service implementation
//   - auditService: Service for auditing user actions
//
// Returns:
//   - *Handler: Initialized gRPC handler
func NewHandler(
	cfg config.Config,
	logger *logger.Logger,
	authorizer service.Authorizer,
	shortener service.Shortener,
	auditService service.AuditService,
) *Handler {
	return &Handler{
		cfg:          cfg,
		logger:       logger,
		authorizer:   authorizer,
		shortener:    shortener,
		auditService: auditService,
	}
}

// ShortenURL handles gRPC requests to shorten URLs.
// Accepts a URL string and returns the shortened version.
//
// Authentication:
//   - Requires Bearer token in Authorization metadata
//
// Parameters:
//   - ctx: Context containing request metadata and cancellation signals
//   - req: URLShortenRequest containing the original URL
//
// Returns:
//   - *URLShortenResponse: Contains the shortened URL
//   - error: gRPC status error if operation fails
//
// Status Codes:
//   - OK (0): Short URL successfully created
//   - ALREADY_EXISTS (6): URL was already shortened previously
//   - INVALID_ARGUMENT (3): Invalid URL format or missing URL
//   - UNAUTHENTICATED (16): Missing or invalid authentication token
//   - INTERNAL (13): Internal server error
//
// Example request:
//
//	rpc ShortenURL(URLShortenRequest) returns (URLShortenResponse);
//	message URLShortenRequest { string url = 1; }
//
// Example response:
//
//	message URLShortenResponse { string result = 1; }
func (h *Handler) ShortenURL(ctx context.Context, req *grpcshortener.URLShortenRequest) (*grpcshortener.URLShortenResponse, error) {
	userID, err := h.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	// Validate URL format
	if !h.isValidURL(req.Url) {
		return nil, status.Error(codes.InvalidArgument, "Invalid URL format")
	}

	// Generate short URL identifier
	shortURL, err := h.shortener.GenerateShortURLPart(ctx, userID, req.Url)
	if err != nil {
		var conflictErr *repository.ErrURLConflict
		if errors.As(err, &conflictErr) {
			// Conflict - URL already exists
			fullURL, buildErr := h.buildFullURL(conflictErr.ShortURL)
			if buildErr != nil {
				h.logger.Error("Failed to build full URL", zap.Error(buildErr))
				return nil, status.Error(codes.Internal, "Internal server error")
			}
			return &grpcshortener.URLShortenResponse{Result: fullURL}, nil
		}
		h.logger.Error("Failed to generate short URL", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	// Audit the shortening event
	h.auditEvent(model.ActionShorten, userID, req.Url)

	// Build full shortened URL
	fullURL, err := h.buildFullURL(shortURL)
	if err != nil {
		h.logger.Error("Failed to build full URL", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &grpcshortener.URLShortenResponse{Result: fullURL}, nil
}

// ExpandURL handles gRPC requests to expand short URLs to original URLs.
// Accepts a short URL identifier and returns the original URL.
//
// Authentication:
//   - Optional (audit events won't be recorded if unauthenticated)
//
// Parameters:
//   - ctx: Context containing request metadata and cancellation signals
//   - req: URLExpandRequest containing the short URL identifier
//
// Returns:
//   - *URLExpandResponse: Contains the original URL
//   - error: gRPC status error if operation fails
//
// Status Codes:
//   - OK (0): Original URL successfully retrieved
//   - NOT_FOUND (5): Short URL not found or has been deleted
//   - INVALID_ARGUMENT (3): Missing short URL identifier
//   - INTERNAL (13): Internal server error
//
// Example request:
//
//	rpc ExpandURL(URLExpandRequest) returns (URLExpandResponse);
//	message URLExpandRequest { string id = 1; }
//
// Example response:
//
//	message URLExpandResponse { string result = 1; }
func (h *Handler) ExpandURL(ctx context.Context, req *grpcshortener.URLExpandRequest) (*grpcshortener.URLExpandResponse, error) {
	userID, _ := h.authenticate(ctx) // Authentication is optional for this method

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	// Retrieve original URL by short identifier
	url, err := h.shortener.GetURLByShortURLPart(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get URL by short URL", zap.Error(err), zap.String("id", req.Id))
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	if url.IsDeleted {
		return nil, status.Error(codes.NotFound, "URL has been deleted")
	}

	// Audit the URL access event
	h.auditEvent(model.ActionFollow, userID, url.OriginalURL)

	return &grpcshortener.URLExpandResponse{Result: url.OriginalURL}, nil
}

// ListUserURLs handles gRPC requests to list all URLs shortened by the authenticated user.
// Returns all URLs created by the user with both short and original URLs.
//
// Authentication:
//   - Requires Bearer token in Authorization metadata
//
// Parameters:
//   - ctx: Context containing request metadata and cancellation signals
//   - _: Empty request parameter
//
// Returns:
//   - *UserURLsResponse: Contains list of user's URLs
//   - error: gRPC status error if operation fails
//
// Status Codes:
//   - OK (0): User URLs retrieved successfully
//   - UNAUTHENTICATED (16): Missing or invalid authentication token
//   - INTERNAL (13): Internal server error
//
// Example request:
//
//	rpc ListUserURLs(Empty) returns (UserURLsResponse);
//	message Empty {}
//
// Example response:
//
//	message UserURLsResponse {
//	  repeated URLData url = 1;
//	}
//	message URLData {
//	  string short_url = 1;
//	  string original_url = 2;
//	}
func (h *Handler) ListUserURLs(ctx context.Context, _ *grpcshortener.Empty) (*grpcshortener.UserURLsResponse, error) {
	userID, err := h.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	// Retrieve all URLs created by the user
	userURLs, err := h.shortener.GetURLsByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user URLs", zap.Error(err), zap.String("userID", userID))
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	// Convert to gRPC response format
	response := &grpcshortener.UserURLsResponse{}
	for _, url := range userURLs {
		fullURL, err := h.buildFullURL(url.ShortURL)
		if err != nil {
			h.logger.Error("Failed to build full URL", zap.Error(err))
			continue
		}

		response.Url = append(response.Url, &grpcshortener.URLData{
			ShortUrl:    fullURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return response, nil
}

// authenticate extracts and validates authorization token from gRPC metadata.
// It retrieves the Bearer token from the Authorization header and validates it.
//
// Parameters:
//   - ctx: Context containing gRPC metadata
//
// Returns:
//   - string: User ID extracted from the valid token
//   - error: gRPC status error if authentication fails
//
// Authentication Flow:
//  1. Extract metadata from context
//  2. Retrieve Authorization header
//  3. Parse Bearer token
//  4. Validate token and extract user ID
func (h *Handler) authenticate(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "Authorization metadata is required")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", status.Error(codes.Unauthenticated, "Authorization header is required")
	}

	token := strings.TrimPrefix(authHeaders[0], "Bearer ")
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "Bearer token is required")
	}

	// Validate token and extract user ID
	userID, err := h.authorizer.ValidateToken(token)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "Invalid token")
	}

	return userID, nil
}

// isValidURL validates URL format by checking for proper protocol prefix.
// It ensures URLs start with either http:// or https://.
//
// Parameters:
//   - urlStr: URL string to validate
//
// Returns:
//   - bool: true if URL is valid, false otherwise
func (h *Handler) isValidURL(urlStr string) bool {
	return strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://")
}

// buildFullURL constructs a complete shortened URL from base URL and short identifier.
// It properly handles trailing slashes in the base URL configuration.
//
// Parameters:
//   - shortURL: Short URL identifier part
//
// Returns:
//   - string: Complete shortened URL
//   - error: If URL construction fails
func (h *Handler) buildFullURL(shortURL string) (string, error) {
	if strings.HasSuffix(h.cfg.BaseURL, "/") {
		return h.cfg.BaseURL + shortURL, nil
	}
	return h.cfg.BaseURL + "/" + shortURL, nil
}

// auditEvent sends an audit event to the audit service.
// It creates an audit event with timestamp, action, user ID, and URL.
//
// Parameters:
//   - action: Type of audit action (shorten, follow, etc.)
//   - userID: User ID who performed the action
//   - url: URL associated with the action
func (h *Handler) auditEvent(action model.AuditAction, userID, url string) {
	if h.auditService == nil {
		return
	}

	event := model.NewAuditEvent(time.Now().Unix(), action, userID, url)
	h.auditService.NotifyAll(*event)
}

// RegisterService registers the gRPC handler with a gRPC server.
// This method implements the gRPC service registration pattern.
//
// Parameters:
//   - server: gRPC server instance to register with
func (h *Handler) RegisterService(server *grpc.Server) {
	grpcshortener.RegisterShortenerServiceServer(server, h)
}

// UnaryInterceptorMiddleware creates a gRPC unary interceptor for logging and observability.
// It logs incoming requests and outgoing responses for debugging and monitoring.
//
// Returns:
//   - grpc.UnaryServerInterceptor: Interceptor function for gRPC unary calls
//
// Interceptor Capabilities:
//   - Request logging with method name and payload
//   - Response logging with result or error
//   - Error logging for failed requests
func (h *Handler) UnaryInterceptorMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Log incoming request
		h.logger.Debug("gRPC request",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
		)

		// Execute the handler
		resp, err := handler(ctx, req)

		// Log result
		if err != nil {
			h.logger.Error("gRPC handler error",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
		} else {
			h.logger.Debug("gRPC response",
				zap.String("method", info.FullMethod),
				zap.Any("response", resp),
			)
		}

		return resp, err
	}
}
