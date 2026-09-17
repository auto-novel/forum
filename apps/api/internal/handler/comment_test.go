package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"auth/internal/httpx"
	"auth/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type editableCommentRepository struct {
	repository.CommentRepository
	comment repository.Comment
	updated bool
}

func (r *editableCommentRepository) Find(int16, int64) (*repository.Comment, error) {
	return &r.comment, nil
}

func (r *editableCommentRepository) Update(_ int16, _ int64, content string) (*repository.Comment, error) {
	r.updated = true
	r.comment.Content = content
	return &r.comment, nil
}

func TestAdminCanEditCommentAfterWindow(t *testing.T) {
	for _, tc := range []struct {
		name       string
		role       string
		userID     int64
		age        time.Duration
		wantStatus int
	}{
		{"author within window", "member", 1, 19 * time.Minute, http.StatusOK},
		{"author after window", "member", 1, 21 * time.Minute, http.StatusForbidden},
		{"admin after window", "admin", 2, 21 * time.Minute, http.StatusOK},
		{"admin editing own old comment", "admin", 1, 21 * time.Minute, http.StatusOK},
		{"other member within window", "member", 2, 19 * time.Minute, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &editableCommentRepository{comment: repository.Comment{
				ID: 7, SubjectKey: "42", AuthorID: 1, CreatedAt: time.Now().Add(-tc.age),
			}}
			router := chi.NewRouter()
			NewCommentHandler(repo, nil).RegisterRoutes(router)
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": "tester", "uid": tc.userID, "role": tc.role,
			}).SignedString([]byte(httpx.AccessTokenSecret))
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPatch, "/7", strings.NewReader(`{"content":"更新内容"}`))
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != tc.wantStatus || repo.updated != (tc.wantStatus == http.StatusOK) {
				t.Fatalf("status=%d want=%d updated=%v body=%s", recorder.Code, tc.wantStatus, repo.updated, recorder.Body.String())
			}
		})
	}
}

func TestValidateComment(t *testing.T) {
	zero := int64(0)
	positive := int64(1)
	for _, tc := range []struct {
		name    string
		input   commentInput
		wantErr bool
	}{
		{"normal", commentInput{Content: "评论", RootID: &positive}, false},
		{"max length", commentInput{Content: strings.Repeat("字", 1000)}, false},
		{"empty", commentInput{Content: " \n\t"}, true},
		{"too long", commentInput{Content: strings.Repeat("字", 1001)}, true},
		{"too long with padding", commentInput{Content: strings.Repeat("字", 1000) + " "}, true},
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

func TestValidateCommentUpdate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"content only", `{"content":"更新内容"}`, false},
		{"root ID", `{"content":"更新内容","rootId":1}`, true},
		{"null root ID", `{"content":"更新内容","rootId":null}`, true},
		{"blank content", `{"content":" \n\t"}`, true},
		{"long content", `{"content":"` + strings.Repeat("字", 1001) + `"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPatch, "/comment/1", strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			input, err := httpx.Body[commentUpdateInput](request)
			if err == nil {
				err = validateCommentUpdate(input, nil)
			}
			if (err != nil) != tc.wantErr {
				t.Fatalf("comment update error = %v, want error %v", err, tc.wantErr)
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
