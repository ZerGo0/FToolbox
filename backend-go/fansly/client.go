package fansly

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	baseURL        = "https://apiv3.fansly.com/api/v1"
	defaultTimeout = 30 * time.Second
)

type Client struct {
	httpClient *http.Client
	rateLimit  int
	requests   []time.Time
	mu         sync.Mutex
	authToken  string
}

func NewClient(rateLimit int) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		rateLimit: rateLimit,
		requests:  make([]time.Time, 0),
		authToken: getEnv("FANSLY_AUTH_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// checkRateLimit ensures we don't exceed the rate limit
func (c *Client) checkRateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	oneMinuteAgo := now.Add(-time.Minute)

	// Remove requests older than 1 minute
	var recentRequests []time.Time
	for _, reqTime := range c.requests {
		if reqTime.After(oneMinuteAgo) {
			recentRequests = append(recentRequests, reqTime)
		}
	}
	c.requests = recentRequests

	// If we're at the limit, wait
	if len(c.requests) >= c.rateLimit {
		oldestRequest := c.requests[0]
		waitTime := oldestRequest.Add(time.Minute).Sub(now)
		if waitTime > 0 {
			zap.L().Debug("Rate limit reached, waiting", zap.Duration("wait", waitTime))
			time.Sleep(waitTime)
		}
	}

	// Add current request
	c.requests = append(c.requests, now)
}

// doRequest performs an HTTP request with rate limiting
func (c *Client) doRequest(url string) ([]byte, error) {
	c.checkRateLimit()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "ftoolbox (contact: zergo0@pr.mozmail.com")
	if c.authToken != "" {
		req.Header.Set("Authorization", c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
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
	url := fmt.Sprintf("%s/contentdiscovery/media/tag?tag=%s&ngsw-bypass=true", baseURL, tagName)

	body, err := c.doRequest(url)
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
		return nil, fmt.Errorf("tag not found")
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
	url := fmt.Sprintf("%s/contentdiscovery/media/suggestionsnew?before=0&after=0&tagIds=%s&limit=%d&offset=%d&ngsw-bypass=true",
		baseURL, tagID, limit, offset)

	body, err := c.doRequest(url)
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
	url := fmt.Sprintf("%s/contentdiscovery/media/suggestionsnew?before=0&after=%s&tagIds=%s&limit=%d&offset=0&ngsw-bypass=true",
		baseURL, after, tagID, limit)

	body, err := c.doRequest(url)
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
	tag, err := c.GetTag(tagName)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch view count: %w", err)
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
