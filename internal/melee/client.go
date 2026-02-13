package melee

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// defaultHeaders returns browser-like headers to avoid 403 from anti-bot checks.
func defaultHeaders() http.Header {
	return http.Header{
		"User-Agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
		"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		"Accept-Language": {"en-US,en;q=0.9"},
		"Accept-Encoding":  {"gzip, deflate"},
		"Connection":      {"keep-alive"},
		"Upgrade-Insecure-Requests": {"1"},
		"Sec-Fetch-Dest":   {"document"},
		"Sec-Fetch-Mode":   {"navigate"},
		"Sec-Fetch-Site":   {"none"},
		"Sec-Fetch-User":   {"?1"},
	}
}

func (c *Client) FetchTournamentDate(meleeID int) (time.Time, error) {
	url := fmt.Sprintf("https://melee.gg/Tournament/View/%d", meleeID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}
	for k, v := range defaultHeaders() {
		req.Header[k] = v
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to fetch tournament page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("tournament page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read response body: %w", err)
	}

	html := string(body)

	// Look for <span data-toggle="datetime" data-value="8/31/2024 7:00:00 AM">
	re := regexp.MustCompile(`data-toggle="datetime"\s+data-value="([^"]+)"`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("could not find tournament date in page")
	}

	dateStr := strings.TrimSpace(matches[1])

	// Try multiple date formats
	formats := []string{
		"1/2/2006 3:04:05 PM",
		"1/2/2006 3:04:05 AM",
		"01/02/2006 3:04:05 PM",
		"01/02/2006 3:04:05 AM",
		"1/2/2006 3:04 PM",
		"1/2/2006 3:04 AM",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}
