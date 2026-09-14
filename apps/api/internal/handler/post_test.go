package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/repository"
)

type listPostRepository struct {
	repository.PostRepository
	items []repository.PostDetails
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
