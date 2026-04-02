package util_test

import (
	"testing"

	"github.com/akyrey/openapi-splitter/internal/util"
)

func TestNormalizePathForFileName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple path",
			input: "/pets",
			want:  "pets",
		},
		{
			name:  "path with parameter",
			input: "/pets/{petId}",
			want:  "pets_petId",
		},
		{
			name:  "nested path",
			input: "/api/v1/users",
			want:  "api_v1_users",
		},
		{
			name:  "path with multiple parameters",
			input: "/users/{userId}/posts/{postId}",
			want:  "users_userId_posts_postId",
		},
		{
			name:  "path with hyphen",
			input: "/my-resource/{id}",
			want:  "my_resource_id",
		},
		{
			name:  "root path",
			input: "/",
			want:  "",
		},
		{
			name:  "path with dots",
			input: "/v1.2/resource",
			want:  "v1_2_resource",
		},
		{
			name:  "no leading slash",
			input: "pets",
			want:  "pets",
		},
		{
			name:  "path with query-like chars",
			input: "/items?filter",
			want:  "items_filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := util.NormalizePathForFileName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePathForFileName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
