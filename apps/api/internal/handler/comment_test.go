package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateComment(t *testing.T) {
	zero := int64(0)
	positive := int64(1)
	for _, tc := range []struct {
		name    string
		input   commentInput
		wantErr bool
	}{
		{"normal", commentInput{Content: "评论", RootID: &positive}, false},
		{"max length", commentInput{Content: strings.Repeat("字", 100000)}, false},
		{"empty", commentInput{Content: " \n\t"}, true},
		{"too long", commentInput{Content: strings.Repeat("字", 100001)}, true},
		{"too long with padding", commentInput{Content: strings.Repeat("字", 100000) + " "}, true},
		{"invalid root", commentInput{Content: "评论", RootID: &zero}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateComment(tc.input, nil)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateComment() error = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}

func TestCommentResponsesMaskModeratedContent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int16
		content string
	}{
		{"published", repository.StatusPublished, "原始内容"},
		{"hidden", repository.StatusHidden, ""},
		{"deleted", repository.StatusDeleted, ""},
		{"unknown", 99, ""},
	} {
		for _, role := range []string{"", "member", "admin"} {
			t.Run(tc.name+"/"+role, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, "/", nil)
				if role != "" {
					token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
						"sub": "tester", "uid": 1, "role": role,
					}).SignedString([]byte(httpx.AccessTokenSecret))
					if err != nil {
						t.Fatal(err)
					}
					request.Header.Set("Authorization", "Bearer "+token)
				}
				recorder := httptest.NewRecorder()
				called := false
				httpx.OptionalAccessToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
					expected := tc.content
					if role == "admin" {
						expected = "原始内容"
					}
					rootID := int64(1)
					comment := repository.Comment{
						ID: 2, SubjectKey: "123", RootID: &rootID,
						Content: "原始内容", Status: tc.status,
					}
					post, err := newCommentResponse(r, comment)
					if err != nil {
						t.Fatal(err)
					}
					external := newExternalCommentResponse(r, comment)
					if post.Content != expected || external.Content != expected {
						t.Fatalf("unexpected content: post=%q external=%q", post.Content, external.Content)
					}
					if post.Status != tc.status || external.Status != tc.status {
						t.Fatal("comment status was not preserved")
					}
					if post.ID != comment.ID || external.ID != comment.ID ||
						post.RootID == nil || external.RootID == nil ||
						*post.RootID != rootID || *external.RootID != rootID {
						t.Fatal("comment identity or reply relationship was not preserved")
					}
				})).ServeHTTP(recorder, request)
				if !called {
					t.Fatalf("authentication failed: %d", recorder.Code)
				}
			})
		}
	}
}
