package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/username/otel-playground/internal/lib/collections"
	libjson "github.com/username/otel-playground/internal/lib/json"
)

var successfulStatuses = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNonAuthoritativeInfo,
	http.StatusNoContent,
}

type ResponseError http.Response

// Error fulfills the error interface.
func (se *ResponseError) Error() string {
	return fmt.Sprintf("response error for %s", se.Request.URL.Redacted())
}

func GetJSON(url string, dst interface{}) error {
	c := http.Client{
		Timeout: time.Duration(1) * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request %w", err)
	}

	req.Header.Add("Accept", `application/json`)
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get url '%s' %w", url, err)
	}
	defer res.Body.Close()

	if !collections.SliceContains(res.StatusCode, successfulStatuses) {
		return fmt.Errorf("%w: unexpected status: %d", (*ResponseError)(res), res.StatusCode)
	}

	if err := libjson.Decode(res.Body, dst); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	return nil
}
