package xweb

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

// the func is the same as condition ? true : false
func Ternary(express bool, trueVal interface{}, falseVal interface{}) interface{} {
	if express {
		return trueVal
	}
	return falseVal
}

// internal utility methods
func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

func JoinPath(paths ...string) string {
	if len(paths) < 1 {
		return ""
	}
	res := ""
	for _, p := range paths {
		res = path.Join(res, p)
	}
	return res
}

func PageSize(total, limit int) int {
	if total <= 0 {
		return 1
	} else {
		x := total % limit
		if x > 0 {
			return total/limit + 1
		} else {
			return total / limit
		}
	}
}

func SimpleParse(data string) map[string]string {
	configs := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		vs := strings.Split(line, "=")
		if len(vs) == 2 {
			configs[strings.TrimSpace(vs[0])] = strings.TrimSpace(vs[1])
		}
	}
	return configs
}

func dirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}

	return true
}

func fileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// Urlencode is a helper method that converts a map into URL-encoded form data.
// It is a useful when constructing HTTP POST requests.
func Urlencode(data map[string]string) string {
	var buf bytes.Buffer
	for k, v := range data {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
		buf.WriteByte('&')
	}
	s := buf.String()
	return s[0 : len(s)-1]
}

func UnTitle(s string) string {
	if len(s) < 2 {
		return strings.ToLower(s)
	}
	return strings.ToLower(string(s[0])) + s[1:]
}

var slugRegex = regexp.MustCompile(`(?i:[^a-z0-9\-_])`)

// Slug is a helper function that returns the URL slug for string s.
// It's used to return clean, URL-friendly strings that can be
// used in routing.
func Slug(s string, sep string) string {
	if s == "" {
		return ""
	}
	slug := slugRegex.ReplaceAllString(s, sep)
	if slug == "" {
		return ""
	}
	quoted := regexp.QuoteMeta(sep)
	sepRegex := regexp.MustCompile("(" + quoted + "){2,}")
	slug = sepRegex.ReplaceAllString(slug, sep)
	sepEnds := regexp.MustCompile("^" + quoted + "|" + quoted + "$")
	slug = sepEnds.ReplaceAllString(slug, "")
	return strings.ToLower(slug)
}

// NewCookie is a helper method that returns a new http.Cookie object.
// Duration is specified in seconds. If the duration is zero, the cookie is permanent.
// This can be used in conjunction with ctx.SetCookie.
func NewCookie(name string, value string, age int64) *http.Cookie {
	var utctime time.Time
	if age == 0 {
		// 2^31 - 1 seconds (roughly 2038)
		utctime = time.Unix(2147483647, 0)
	} else {
		utctime = time.Unix(time.Now().Unix()+age, 0)
	}
	return &http.Cookie{Name: name, Value: value, Expires: utctime}
}

func removeStick(uri string) string {
	uri = strings.TrimRight(uri, "/")
	if uri == "" {
		uri = "/"
	}
	return uri
}

var (
	fieldCache      = make(map[reflect.Type]map[string]int)
	fieldCacheMutex sync.RWMutex
)

// this method cache fields' index to field name
func fieldByName(v reflect.Value, name string) reflect.Value {
	t := v.Type()
	fieldCacheMutex.RLock()
	cache, ok := fieldCache[t]
	fieldCacheMutex.RUnlock()
	if !ok {
		cache = make(map[string]int)
		for i := 0; i < v.NumField(); i++ {
			cache[t.Field(i).Name] = i
		}
		fieldCacheMutex.Lock()
		fieldCache[t] = cache
		fieldCacheMutex.Unlock()
	}

	if i, ok := cache[name]; ok {
		return v.Field(i)
	}

	return reflect.Zero(t)
}
