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
	"strings"
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
	Success  bool `json:"success"`
	Response any  `json:"response,omitempty"`
}

// FanslyTag represents a tag from the Fansly API
type FanslyTag struct {
	ID          string `json:"id"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
	ViewCount   int64  `json:"viewCount"`
	PostCount   int64  `json:"postCount"`
	Flags       int    `json:"flags"`
	CreatedAt   int64  `json:"createdAt"`
}

// TagResponseData represents the tag response data structure
type TagResponseData struct {
	MediaOfferSuggestionTag *FanslyTag     `json:"mediaOfferSuggestionTag,omitempty"`
	AggregationData         map[string]any `json:"aggregationData,omitempty"`
}

// GetTagWithContext fetches a single tag by name with context
func (c *Client) GetTagWithContext(ctx context.Context, tagName string) (*TagResponseData, error) {
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

	return response.Response, nil
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
	ReplyPermissionFlags any                `json:"replyPermissionFlags,omitempty"`
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

// FanslyAccount represents a Fansly account/creator
type FanslyAccount struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	DisplayName       string `json:"displayName"`
	AccountMediaLikes int64  `json:"accountMediaLikes"`
	PostLikes         int64  `json:"postLikes"`
	FollowCount       int64  `json:"followCount"`
	TimelineStats     struct {
		ImageCount int64 `json:"imageCount"`
		VideoCount int64 `json:"videoCount"`
	} `json:"timelineStats"`
}

// SuggestionsResponseData represents the complete response data structure from suggestionsnew endpoint
type SuggestionsResponseData struct {
	MediaOfferSuggestions []MediaOfferSuggestion `json:"mediaOfferSuggestions,omitempty"`
	AggregationData       *struct {
		Accounts            []FanslyAccount `json:"accounts,omitempty"`
		AccountMedia        []any           `json:"accountMedia,omitempty"`
		AccountMediaBundles []any           `json:"accountMediaBundles,omitempty"`
		Posts               []FanslyPost    `json:"posts,omitempty"`
		Tips                []any           `json:"tips,omitempty"`
		TipGoals            []any           `json:"tipGoals,omitempty"`
		Stories             []any           `json:"stories,omitempty"`
	} `json:"aggregationData,omitempty"`
}

// GetSuggestionsData fetches all data from the suggestions endpoint with full parameter control
func (c *Client) GetSuggestionsData(ctx context.Context, tagIDs []string, before, after string, limit, offset int) (*SuggestionsResponseData, error) {
	// Build tag IDs string
	var tagIDsStr strings.Builder
	for i, id := range tagIDs {
		if i > 0 {
			tagIDsStr.WriteString(",")
		}
		tagIDsStr.WriteString(id)
	}

	url := fmt.Sprintf("%s/contentdiscovery/media/suggestionsnew?before=%s&after=%s&tagIds=%s&limit=%d&offset=%d&ngsw-bypass=true",
		baseURL, before, after, tagIDsStr.String(), limit, offset)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Success  bool                     `json:"success"`
		Response *SuggestionsResponseData `json:"response"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Response, nil
}

// GetAccountsWithContext fetches multiple accounts by their IDs with context
func (c *Client) GetAccountsWithContext(ctx context.Context, accountIDs []string) ([]FanslyAccount, error) {
	if len(accountIDs) == 0 {
		return nil, nil
	}

	// Build comma-separated list of account IDs
	var idsParam strings.Builder
	for i, id := range accountIDs {
		if i > 0 {
			idsParam.WriteString(",")
		}
		idsParam.WriteString(id)
	}

	url := fmt.Sprintf("%s/account?ids=%s&ngsw-bypass=true", baseURL, idsParam.String())

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Success  bool            `json:"success"`
		Response []FanslyAccount `json:"response"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return response.Response, nil
}

// GetAccountByUsername fetches a single account by username
func (c *Client) GetAccountByUsername(ctx context.Context, username string) (*FanslyAccount, error) {
	url := fmt.Sprintf("%s/account?usernames=%s&ngsw-bypass=true", baseURL, username)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	rawResponse := string(body)
	zap.L().Info("Raw response", zap.String("response", rawResponse))

	var response struct {
		Success  bool            `json:"success"`
		Response []FanslyAccount `json:"response"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	if len(response.Response) == 0 {
		return nil, fmt.Errorf("account not found")
	}

	return &response.Response[0], nil
}

// ParseFanslyTimestamp converts Fansly timestamp (milliseconds) to time.Time
func ParseFanslyTimestamp(timestamp any) time.Time {
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
