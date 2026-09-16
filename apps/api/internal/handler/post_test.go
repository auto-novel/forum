package handler

import (
	"auth/internal/httpx"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth/internal/repository"
)

type listPostRepository struct {
	repository.PostRepository
	items []repository.PostDetails
}

func TestValidatePostTextLimits(t *testing.T) {
	handler := &postHandler{}
	for _, tc := range []struct {
		name    string
		title   string
		content string
		wantErr bool
	}{
		{"minimum lengths", "标题", "文", false},
		{"maximum lengths", strings.Repeat("题", 100), strings.Repeat("文", 20000), false},
		{"short title", "题", "正文", true},
		{"long title", strings.Repeat("题", 101), "正文", true},
		{"blank title", " \n\t", "正文", true},
		{"empty content", "标题", "", true},
		{"blank content", "标题", " \n\t", true},
		{"long content", "标题", strings.Repeat("文", 20001), true},
		{"long content with padding", "标题", strings.Repeat("文", 20000) + " ", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := handler.validatePost(postInput{CategoryID: 1, Title: tc.title, Content: tc.content})
			if (err != nil) != tc.wantErr {
				t.Fatalf("validatePost() error = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}

func (r listPostRepository) List(repository.PostFilter, int64, int64) (int64, []repository.PostDetails, error) {
	return int64(len(r.items)), r.items, nil
}

func TestPostListOmitsContentAndDetailPreservesIt(t *testing.T) {
	post := repository.PostDetails{Post: repository.Post{
		ID: 42, Title: "标题", Content: "完整正文", AuthorUsername: "alice", CommentsCount: 3,
	}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/post/", nil)
	if err := respondPosts(recorder, request, listPostRepository{items: []repository.PostDetails{post}}, nil, repository.PostFilter{}); err != nil {
		t.Fatal(err)
	}
	var response page[map[string]json.RawMessage]
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("unexpected list: %s", recorder.Body.String())
	}
	item := response.Items[0]
	if _, ok := item["content"]; ok {
		t.Fatal("list must not include content, even as an empty string")
	}
	detailJSON, err := json.Marshal(newPostResponse(post, false))
	if err != nil {
		t.Fatal(err)
	}
	var detail map[string]json.RawMessage
	if err := json.Unmarshal(detailJSON, &detail); err != nil {
		t.Fatal(err)
	}
	var content string
	if err := json.Unmarshal(detail["content"], &content); err != nil {
		t.Fatal(err)
	}
	if content != post.Content {
		t.Fatalf("detail content = %q", content)
	}
	delete(detail, "content")
	if len(item) != len(detail) {
		t.Fatal("list metadata fields differ from detail")
	}
	for key, value := range detail {
		if string(item[key]) != string(value) {
			t.Fatalf("list metadata %s differs from detail", key)
		}
	}
}

type writePostRepository struct {
	repository.PostRepository
	written bool
}

func (r *writePostRepository) Find(int64, bool) (*repository.PostDetails, error) {
	return &repository.PostDetails{Post: repository.Post{ID: 42, AuthorID: 1}}, nil
}

func (r *writePostRepository) Create(input repository.CreatePostInput) (*repository.PostDetails, error) {
	r.written = true
	return &repository.PostDetails{Post: repository.Post{ID: 42, CategoryID: input.CategoryID}}, nil
}

func (r *writePostRepository) Update(id int64, input repository.UpdatePostInput) (*repository.PostDetails, error) {
	r.written = true
	return &repository.PostDetails{Post: repository.Post{ID: id, CategoryID: input.CategoryID}}, nil
}

type noFavoriteRepository struct{ repository.FavoriteRepository }

func (noFavoriteRepository) Has(int64, int64) (bool, error) { return false, nil }

func TestAnnouncementsPublishingRequiresAdmin(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPatch} {
		for _, role := range []string{"member", "trusted", "admin"} {
			for _, categoryID := range []int64{1, 2, 3} {
				t.Run(fmt.Sprintf("%s/%s/%d", method, role, categoryID), func(t *testing.T) {
					repo := &writePostRepository{}
					router := chi.NewRouter()
					NewPostHandler(repo, noFavoriteRepository{}, nil, nil).RegisterRoutes(router)
					path := "/"
					wantStatus := http.StatusCreated
					if method == http.MethodPatch {
						path = "/42/"
						wantStatus = http.StatusOK
					}
					allowed := categoryID != 2 || role == "admin"
					if !allowed {
						wantStatus = http.StatusForbidden
					}
					request := httptest.NewRequest(method, path, strings.NewReader(fmt.Sprintf(
						`{"categoryId":%d,"title":"标题","content":"正文","tagIds":[]}`, categoryID)))
					request.Header.Set("Content-Type", "application/json")
					token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
						"sub": "tester", "uid": 1, "role": role,
					}).SignedString([]byte(httpx.AccessTokenSecret))
					if err != nil {
						t.Fatal(err)
					}
					request.Header.Set("Authorization", "Bearer "+token)
					recorder := httptest.NewRecorder()
					router.ServeHTTP(recorder, request)
					if recorder.Code != wantStatus || repo.written != allowed {
						t.Fatalf("status=%d want=%d written=%v allowed=%v body=%s", recorder.Code, wantStatus, repo.written, allowed, recorder.Body.String())
					}
				})
			}
		}
	}
}
