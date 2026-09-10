//go:build integration

package tests

import (
	"auth/internal/repository"
	"errors"
	"testing"
)

func TestJetRepositories(t *testing.T) {
	resetDatabase()
	category, err := categoryRepo.Create("general", nil, `{}`)
	if err != nil {
		t.Fatal(err)
	}
	tag, err := tagRepo.Create(category.ID, "公告", 1, 10, `{}`)
	if err != nil {
		t.Fatal(err)
	}

	post, err := postRepo.Create(repository.CreatePostInput{
		CategorySlug: category.Slug,
		Title:        "第一篇帖子", Content: "正文", AuthorID: 7, AuthorUsername: "alice",
		TagIDs: []int64{tag.ID}, Attr: `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(post.Tags) != 1 || post.Tags[0].ID != tag.ID {
		t.Fatalf("unexpected tags: %#v", post.Tags)
	}
	const otherSubjectType int16 = 1
	otherSubjectComment, err := commentRepo.Create(repository.CreateCommentInput{
		SubjectType:    otherSubjectType,
		SubjectKey:     "novel:chapter-1",
		Content:        "其他主体评论",
		AuthorID:       8,
		AuthorUsername: "bob",
		Attr:           `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	otherTotal, otherComments, err := commentRepo.List(otherSubjectType, "novel:chapter-1", 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if otherTotal != 1 || len(otherComments) != 1 || otherComments[0].SubjectKey != "novel:chapter-1" {
		t.Fatalf("unexpected external comments: total=%d items=%#v", otherTotal, otherComments)
	}
	var commentsCount int32
	if err := testDB.QueryRow("SELECT comments_count FROM post WHERE id = $1", post.ID).Scan(&commentsCount); err != nil {
		t.Fatal(err)
	}
	if commentsCount != 0 {
		t.Fatalf("post comments count changed after creating a non-post comment: %d", commentsCount)
	}
	if err := commentRepo.SetStatus(otherSubjectType, otherSubjectComment.ID, repository.StatusDeleted); err != nil {
		t.Fatal(err)
	}
	if err := testDB.QueryRow("SELECT comments_count FROM post WHERE id = $1", post.ID).Scan(&commentsCount); err != nil {
		t.Fatal(err)
	}
	if commentsCount != 0 {
		t.Fatalf("post comments count changed after moderating a non-post comment: %d", commentsCount)
	}

	total, posts, err := postRepo.List(repository.PostFilter{CategorySlug: category.Slug, TagIDs: []int64{tag.ID}}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(posts) != 1 {
		t.Fatalf("unexpected posts: total=%d items=%d", total, len(posts))
	}

	viewed, err := postRepo.Find(post.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if viewed.ViewsCount != 1 {
		t.Fatalf("views count = %d", viewed.ViewsCount)
	}

	root, err := commentRepo.Create(repository.CreateCommentInput{
		SubjectType:    repository.CommentSubjectPost,
		SubjectKey:     repository.PostSubjectKey(post.ID),
		Content:        "评论",
		AuthorID:       8,
		AuthorUsername: "bob",
		Attr:           `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if root.SubjectType != repository.CommentSubjectPost || root.SubjectKey != repository.PostSubjectKey(post.ID) {
		t.Fatalf("unexpected comment subject: type=%d key=%s", root.SubjectType, root.SubjectKey)
	}
	_, err = commentRepo.Create(repository.CreateCommentInput{
		SubjectType:    repository.CommentSubjectPost,
		SubjectKey:     repository.PostSubjectKey(post.ID),
		RootID:         &root.ID,
		Content:        "回复",
		AuthorID:       7,
		AuthorUsername: "alice",
		Attr:           `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	commentTotal, comments, err := commentRepo.List(repository.CommentSubjectPost, repository.PostSubjectKey(post.ID), 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if commentTotal != 2 || len(comments) != 2 {
		t.Fatalf("unexpected comments: total=%d items=%d", commentTotal, len(comments))
	}

	if err := favoriteRepo.Set(post.ID, 7, true); err != nil {
		t.Fatal(err)
	}
	favorited, err := favoriteRepo.Has(post.ID, 7)
	if err != nil {
		t.Fatal(err)
	}
	if !favorited {
		t.Fatal("favorite was not found after insertion")
	}
	favoriteTotal, _, err := postRepo.List(repository.PostFilter{FavoriteUserID: 7}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if favoriteTotal != 1 {
		t.Fatalf("favorite total = %d", favoriteTotal)
	}

	if err := postRepo.SetModeration(post.ID, repository.StatusPublished, true, nil); err != nil {
		t.Fatal(err)
	}
	_, err = commentRepo.Create(repository.CreateCommentInput{
		SubjectType:    repository.CommentSubjectPost,
		SubjectKey:     repository.PostSubjectKey(post.ID),
		Content:        "blocked",
		AuthorID:       9,
		AuthorUsername: "carol",
		Attr:           `{}`,
	})
	if !errors.Is(err, repository.ErrCommentsLocked) {
		t.Fatalf("got %v, want ErrCommentsLocked", err)
	}
	if err := commentRepo.SetStatus(repository.CommentSubjectPost, root.ID, repository.StatusDeleted); err != nil {
		t.Fatal(err)
	}
	updated, err := postRepo.Update(post.ID, repository.UpdatePostInput{
		Title:   "更新标题",
		Content: "更新正文",
		TagIDs:  []int64{tag.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "更新标题" || updated.CommentsCount != 1 {
		t.Fatalf("unexpected updated post: %#v", updated.Post)
	}
}
