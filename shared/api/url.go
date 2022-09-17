package api

import (
	"encoding/json"
	"net/url"
	"strings"
)

// URL represents an endpoint for the LXD API.
type URL struct {
	url.URL
}

// NewURL creates a new URL.
func NewURL() *URL {
	return &URL{}
}

// MarshalJSON implements json.Marshaler for the URL type.
func (u URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// MarshalYAML implements yaml.Marshaler for the URL type. Note that for yaml we just need to return a yaml
// marshallable type (if we return []byte as in the json implementation it is written as an int array).
func (u URL) MarshalYAML() (any, error) {
	return u.String(), nil
}

// UnmarshalJSON implements json.Unmarshaler for URL.
func (u *URL) UnmarshalJSON(b []byte) error {
	var urlStr string
	err := json.Unmarshal(b, &urlStr)
	if err != nil {
		return err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	*u = URL{URL: *url}

	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for URL.
func (u *URL) UnmarshalYAML(unmarshal func(v any) error) error {
	var urlStr string
	err := unmarshal(&urlStr)
	if err != nil {
		return err
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	*u = URL{URL: *url}

	return nil
}

// Scheme sets the scheme of the URL.
func (u *URL) Scheme(scheme string) *URL {
	u.URL.Scheme = scheme

	return u
}

// Host sets the host of the URL.
func (u *URL) Host(host string) *URL {
	u.URL.Host = host

	return u
}

// Path sets the path of the URL from one or more path parts.
// It appends each of the pathParts (escaped using url.PathEscape) prefixed with "/" to the URL path.
func (u *URL) Path(pathParts ...string) *URL {
	var b strings.Builder

	for _, pathPart := range pathParts {
		b.WriteString("/") // Build an absolute URL.
		b.WriteString(url.PathEscape(pathPart))
	}

	u.URL.Path = b.String()

	return u
}

// Project sets the "project" query parameter in the URL if the projectName is not empty or "default".
func (u *URL) Project(projectName string) *URL {
	if projectName != "default" && projectName != "" {
		queryArgs := u.Query()
		queryArgs.Add("project", projectName)
		u.RawQuery = queryArgs.Encode()
	}

	return u
}

// Target sets the "target" query parameter in the URL if the clusterMemberName is not empty or "default".
func (u *URL) Target(clusterMemberName string) *URL {
	if clusterMemberName != "" && clusterMemberName != "none" {
		queryArgs := u.Query()
		queryArgs.Add("target", clusterMemberName)
		u.RawQuery = queryArgs.Encode()
	}

	return u
}

// WithQuery adds a given query parameter with its value to the URL.
func (u *URL) WithQuery(key string, value string) *URL {
	queryArgs := u.Query()
	queryArgs.Add(key, value)
	u.RawQuery = queryArgs.Encode()

	return u
}
