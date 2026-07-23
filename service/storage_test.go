package service

import (
	"os"
	"testing"
)

func withEnv(t *testing.T, kv map[string]string, fn func()) {
	t.Helper()
	saved := map[string]string{}
	for k := range kv {
		saved[k] = os.Getenv(k)
	}
	for k, v := range kv {
		os.Setenv(k, v)
	}
	defer func() {
		for k, v := range saved {
			os.Setenv(k, v)
		}
	}()
	fn()
}

func TestIsS3Backend(t *testing.T) {
	withEnv(t, map[string]string{"STORAGE_BACKEND": ""}, func() {
		if isS3Backend() {
			t.Fatal("expected local backend when STORAGE_BACKEND is unset")
		}
	})
	withEnv(t, map[string]string{"STORAGE_BACKEND": "S3"}, func() {
		if !isS3Backend() {
			t.Fatal("expected s3 backend to be case-insensitive")
		}
	})
}

func TestS3PublicURLDefaultAWS(t *testing.T) {
	withEnv(t, map[string]string{
		"S3_BUCKET":         "my-bucket",
		"S3_REGION":         "eu-west-2",
		"S3_ENDPOINT":       "",
		"S3_USE_PATH_STYLE": "",
		"S3_PUBLIC_URL_BASE": "",
		"S3_KEY_PREFIX":     "",
	}, func() {
		got := s3PublicURL("My Podcast/episode one.mp3")
		want := "https://my-bucket.s3.eu-west-2.amazonaws.com/My%20Podcast/episode%20one.mp3"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestS3PublicURLPathStyle(t *testing.T) {
	withEnv(t, map[string]string{
		"S3_BUCKET":         "my-bucket",
		"S3_REGION":         "us-east-1",
		"S3_ENDPOINT":       "",
		"S3_USE_PATH_STYLE": "true",
		"S3_PUBLIC_URL_BASE": "",
		"S3_KEY_PREFIX":     "",
	}, func() {
		got := s3PublicURL("show/ep.mp3")
		want := "https://s3.us-east-1.amazonaws.com/my-bucket/show/ep.mp3"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestS3PublicURLCustomEndpointVirtualHosted(t *testing.T) {
	withEnv(t, map[string]string{
		"S3_BUCKET":         "spaces-bucket",
		"S3_ENDPOINT":       "https://nyc3.digitaloceanspaces.com",
		"S3_USE_PATH_STYLE": "",
		"S3_PUBLIC_URL_BASE": "",
		"S3_KEY_PREFIX":     "",
	}, func() {
		got := s3PublicURL("show/ep.mp3")
		want := "https://spaces-bucket.nyc3.digitaloceanspaces.com/show/ep.mp3"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestS3PublicURLCustomEndpointWithoutScheme(t *testing.T) {
	withEnv(t, map[string]string{
		"S3_BUCKET":          "spaces-bucket",
		"S3_ENDPOINT":        "nyc3.digitaloceanspaces.com",
		"S3_USE_PATH_STYLE":  "",
		"S3_PUBLIC_URL_BASE": "",
		"S3_KEY_PREFIX":      "",
	}, func() {
		got := s3PublicURL("show/ep.mp3")
		want := "https://spaces-bucket.nyc3.digitaloceanspaces.com/show/ep.mp3"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestS3EndpointHostPortWithoutScheme(t *testing.T) {
	withEnv(t, map[string]string{"S3_ENDPOINT": "minio:9000"}, func() {
		got := s3Endpoint()
		want := "https://minio:9000"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestS3PublicURLOverrideAndPrefix(t *testing.T) {
	withEnv(t, map[string]string{
		"S3_PUBLIC_URL_BASE": "https://cdn.example.com/",
		"S3_KEY_PREFIX":      "/podcasts/",
	}, func() {
		got := s3PublicURL("show/ep.mp3")
		want := "https://cdn.example.com/podcasts/show/ep.mp3"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestGetPublicFileURLEmptyForLocalBackend(t *testing.T) {
	withEnv(t, map[string]string{"STORAGE_BACKEND": "local"}, func() {
		if got := GetPublicFileURL("/assets/show/ep.mp3"); got != "" {
			t.Fatalf("expected empty public URL for local backend, got %q", got)
		}
	})
}

func TestComputeLocationLocalBackend(t *testing.T) {
	dataDir := t.TempDir()
	withEnv(t, map[string]string{"STORAGE_BACKEND": "local", "DATA": dataDir}, func() {
		got := computeLocation("My Show!", "images", "ep 1.jpg")
		want := dataDir + "/My Show-/images/ep 1.jpg"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestComputeLocationS3Backend(t *testing.T) {
	withEnv(t, map[string]string{"STORAGE_BACKEND": "s3"}, func() {
		got := computeLocation("My Show!", "images", "ep 1.jpg")
		want := "My Show-/images/ep 1.jpg"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}
