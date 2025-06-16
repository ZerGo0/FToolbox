package fansly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"ftoolbox/ratelimit"
	"go.uber.org/zap"
)

const (
	baseURL        = "https://apiv3.fansly.com/api/v1"
	defaultTimeout = 30 * time.Second
)

// ErrTagNotFound is returned when a tag doesn't exist on Fansly
var ErrTagNotFound = errors.New("fansly: tag not found")

type Client struct {
	httpClient    *http.Client
	globalLimiter *ratelimit.GlobalRateLimiter
	authToken     string
	logger        *zap.Logger
}

func NewClient() *Client {
	logger := zap.L().Named("fansly")

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: defaultTimeout,
	}

	return &Client{
		httpClient: httpClient,
		authToken:  getEnv("FANSLY_AUTH_TOKEN", ""),
		logger:     logger,
	}
}

// SetGlobalRateLimit configures a global rate limiter for all requests
func (c *Client) SetGlobalRateLimit(maxRequests int, windowSeconds int) {
	c.globalLimiter = ratelimit.NewGlobalRateLimiter(maxRequests, windowSeconds, c.logger)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// doRequest performs an HTTP request with global rate limiting and retry logic
func (c *Client) doRequest(ctx context.Context, url string) ([]byte, error) {
	// Apply global rate limiting if configured
	if c.globalLimiter != nil {
		if err := c.globalLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "ftoolbox (contact: zergo0@pr.mozmail.com)")
	if c.authToken != "" {
		req.Header.Set("Authorization", c.authToken)
	}

	// Simple retry logic for common errors
	var resp *http.Response
	var lastErr error
	retryableStatuses := map[int]bool{
		http.StatusTooManyRequests:     true,
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}

	maxRetries := 3
	backoff := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req.Clone(ctx))
		if err != nil {
			lastErr = err
			c.logger.Warn("Request failed",
				zap.String("url", url),
				zap.Int("attempt", attempt+1),
				zap.Error(err))

			if attempt < maxRetries {
				select {
				case <-time.After(backoff):
					backoff *= 2
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
		}

		if retryableStatuses[resp.StatusCode] && attempt < maxRetries {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			c.logger.Warn("Retryable HTTP error",
				zap.String("url", url),
				zap.Int("status", resp.StatusCode),
				zap.Int("attempt", attempt+1))

			// For rate limit errors, wait longer
			if resp.StatusCode == http.StatusTooManyRequests {
				backoff = 30 * time.Second
			}

			select {
			case <-time.After(backoff):
				backoff *= 2
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		break
	}

	if resp == nil {
		return nil, fmt.Errorf("no response received")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return body, nil
}

// FanslyResponse represents the generic API response structure
type FanslyResponse struct {
	Success  bool        `json:"success"`
	Response interface{} `json:"response,omitempty"`
}

// FanslyTag represents a tag from the Fansly API
type FanslyTag struct {
	ID          string `json:"id"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
	ViewCount   int64  `json:"viewCount"`
	Flags       int    `json:"flags"`
	CreatedAt   int64  `json:"createdAt"`
}

// TagResponseData represents the tag response data structure
type TagResponseData struct {
	MediaOfferSuggestionTag *FanslyTag             `json:"mediaOfferSuggestionTag,omitempty"`
	AggregationData         map[string]interface{} `json:"aggregationData,omitempty"`
}

// SearchTags searches for tags by keyword (not implemented in v3 API)
func (c *Client) SearchTags(keyword string) ([]FanslyTag, error) {
	// Note: The v3 API doesn't have a direct search endpoint
	// This would need to be implemented differently
	return nil, fmt.Errorf("tag search not implemented in v3 API")
}

// GetTag fetches a single tag by name
func (c *Client) GetTag(tagName string) (*FanslyTag, error) {
	return c.GetTagWithContext(context.Background(), tagName)
}

// GetTagWithContext fetches a single tag by name with context
func (c *Client) GetTagWithContext(ctx context.Context, tagName string) (*FanslyTag, error) {
	url := fmt.Sprintf("%s/contentdiscovery/media/tag?tag=%s&ngsw-bypass=true", baseURL, tagName)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Success  bool             `json:"success"`
		Response *TagResponseData `json:"response"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success || response.Response == nil || response.Response.MediaOfferSuggestionTag == nil {
		return nil, ErrTagNotFound
	}

	return response.Response.MediaOfferSuggestionTag, nil
}

// FanslyPost represents a post from the Fansly API
type FanslyPost struct {
	ID                   string             `json:"id"`
	AccountID            string             `json:"accountId"`
	Content              string             `json:"content"`
	CreatedAt            int64              `json:"createdAt"`
	Attachments          []FanslyAttachment `json:"attachments"`
	FypFlags             *int               `json:"fypFlags,omitempty"`
	InReplyTo            *string            `json:"inReplyTo,omitempty"`
	InReplyToRoot        *string            `json:"inReplyToRoot,omitempty"`
	ReplyPermissionFlags interface{}        `json:"replyPermissionFlags,omitempty"`
	ExpiresAt            *int64             `json:"expiresAt,omitempty"`
	LikeCount            int                `json:"likeCount,omitempty"`
	ReplyCount           int                `json:"replyCount,omitempty"`
	WallIDs              []string           `json:"wallIds,omitempty"`
	MediaLikeCount       int                `json:"mediaLikeCount,omitempty"`
	TotalTipAmount       int                `json:"totalTipAmount,omitempty"`
	AttachmentTipAmount  int                `json:"attachmentTipAmount,omitempty"`
	TipAmount            int                `json:"tipAmount,omitempty"`
}

// FanslyAttachment represents an attachment in a post
type FanslyAttachment struct {
	PostID      string `json:"postId"`
	Pos         int    `json:"pos"`
	ContentType int    `json:"contentType"`
	ContentID   string `json:"contentId"`
}

// MediaOfferSuggestion represents a media suggestion with tags
type MediaOfferSuggestion struct {
	ID            string      `json:"id"`
	CorrelationID string      `json:"correlationId"`
	PostTags      []FanslyTag `json:"postTags"`
}

// PostsResponseData represents the posts response data structure
type PostsResponseData struct {
	MediaOfferSuggestions []MediaOfferSuggestion `json:"mediaOfferSuggestions,omitempty"`
	AggregationData       *struct {
		Accounts            []interface{} `json:"accounts,omitempty"`
		AccountMedia        []interface{} `json:"accountMedia,omitempty"`
		AccountMediaBundles []interface{} `json:"accountMediaBundles,omitempty"`
		Posts               []FanslyPost  `json:"posts,omitempty"`
		Tips                []interface{} `json:"tips,omitempty"`
		TipGoals            []interface{} `json:"tipGoals,omitempty"`
		Stories             []interface{} `json:"stories,omitempty"`
	} `json:"aggregationData,omitempty"`
}

// GetPostsForTag fetches posts for a specific tag ID
func (c *Client) GetPostsForTag(tagID string, limit int, offset int) ([]FanslyPost, error) {
	return c.GetPostsForTagWithContext(context.Background(), tagID, limit, offset)
}

// GetPostsForTagWithContext fetches posts for a specific tag ID with context
func (c *Client) GetPostsForTagWithContext(ctx context.Context, tagID string, limit int, offset int) ([]FanslyPost, error) {
	url := fmt.Sprintf("%s/contentdiscovery/media/suggestionsnew?before=0&after=0&tagIds=%s&limit=%d&offset=%d&ngsw-bypass=true",
		baseURL, tagID, limit, offset)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Success  bool               `json:"success"`
		Response *PostsResponseData `json:"response"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success || response.Response == nil || response.Response.AggregationData == nil {
		return []FanslyPost{}, nil
	}

	return response.Response.AggregationData.Posts, nil
}

// PostsWithSuggestions contains both posts and media suggestions
type PostsWithSuggestions struct {
	Posts       []FanslyPost
	Suggestions []MediaOfferSuggestion
}

// GetPostsForTagWithPagination fetches posts and media suggestions with pagination support
func (c *Client) GetPostsForTagWithPagination(tagID string, limit int, after string) (*PostsWithSuggestions, error) {
	return c.GetPostsForTagWithPaginationAndContext(context.Background(), tagID, limit, after)
}

// GetPostsForTagWithPaginationAndContext fetches posts and media suggestions with pagination support and context
func (c *Client) GetPostsForTagWithPaginationAndContext(ctx context.Context, tagID string, limit int, after string) (*PostsWithSuggestions, error) {
	url := fmt.Sprintf("%s/contentdiscovery/media/suggestionsnew?before=0&after=%s&tagIds=%s&limit=%d&offset=0&ngsw-bypass=true",
		baseURL, after, tagID, limit)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Success  bool               `json:"success"`
		Response *PostsResponseData `json:"response"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success || response.Response == nil {
		return &PostsWithSuggestions{
			Posts:       []FanslyPost{},
			Suggestions: []MediaOfferSuggestion{},
		}, nil
	}

	result := &PostsWithSuggestions{
		Suggestions: response.Response.MediaOfferSuggestions,
	}

	if response.Response.AggregationData != nil {
		result.Posts = response.Response.AggregationData.Posts
	}

	return result, nil
}

// FetchTagViewCount fetches the current view count for a tag
func (c *Client) FetchTagViewCount(tagName string) (int64, error) {
	return c.FetchTagViewCountWithContext(context.Background(), tagName)
}

// FetchTagViewCountWithContext fetches the current view count for a tag with context
func (c *Client) FetchTagViewCountWithContext(ctx context.Context, tagName string) (int64, error) {
	tag, err := c.GetTagWithContext(ctx, tagName)
	if err != nil {
		return 0, err
	}
	return tag.ViewCount, nil
}

// ParseFanslyTimestamp converts Fansly timestamp (milliseconds) to time.Time
func ParseFanslyTimestamp(timestamp interface{}) time.Time {
	switch v := timestamp.(type) {
	case float64:
		// Convert milliseconds to seconds
		return time.Unix(int64(v/1000), int64((v-float64(int64(v/1000))*1000)*1e6))
	case string:
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			// Convert milliseconds to seconds
			return time.Unix(ts/1000, (ts%1000)*1e6)
		}
	case int64:
		// Convert milliseconds to seconds
		return time.Unix(v/1000, (v%1000)*1e6)
	}
	return time.Now()
}
