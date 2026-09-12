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
	if !tag.IsActive {
		t.Fatal("new tag is inactive")
	}
	if err := tagRepo.SetActive(tag.ID, false); err != nil {
		t.Fatal(err)
	}
	tag, err = tagRepo.Update(tag.ID, "公告", 2, 20)
	if err != nil {
		t.Fatal(err)
	}
	if tag.IsActive || tag.Color != 2 || tag.SortOrder != 20 {
		t.Fatalf("unexpected updated tag: %#v", tag)
	}
	if err := tagRepo.SetActive(tag.ID, true); err != nil {
		t.Fatal(err)
	}
	inactiveTag, err := tagRepo.Create(category.ID, "停用标签", 3, 30, `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := tagRepo.SetActive(inactiveTag.ID, false); err != nil {
		t.Fatal(err)
	}
	activeTags, err := tagRepo.ListActive()
	if err != nil {
		t.Fatal(err)
	}
	if len(activeTags) != 1 || activeTags[0].ID != tag.ID {
		t.Fatalf("unexpected active tags: %#v", activeTags)
	}
	categoryTags, err := tagRepo.ListByCategory(category.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(categoryTags) != 2 || categoryTags[0].ID != tag.ID || categoryTags[1].ID != inactiveTag.ID {
		t.Fatalf("unexpected category tags: %#v", categoryTags)
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
	otherTotal, otherComments, err = commentRepo.List(otherSubjectType, "novel:chapter-1", 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if otherTotal != 1 || len(otherComments) != 1 || otherComments[0].Status != repository.StatusDeleted {
		t.Fatalf("deleted external comment missing: total=%d items=%#v", otherTotal, otherComments)
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
	favoriteIDs, err := favoriteRepo.ListPostIDs(7, []int64{post.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !favoriteIDs[post.ID] {
		t.Fatal("favorite post ID was not returned")
	}
	favoriteTotal, _, err := postRepo.List(repository.PostFilter{FavoriteUserID: 7}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if favoriteTotal != 1 {
		t.Fatalf("favorite total = %d", favoriteTotal)
	}

	if err := postRepo.SetCommentsLocked(post.ID, true); err != nil {
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
	for _, status := range []int16{repository.StatusHidden, repository.StatusDeleted} {
		if err := commentRepo.SetStatus(repository.CommentSubjectPost, root.ID, status); err != nil {
			t.Fatal(err)
		}
		for offset := int64(0); offset < 2; offset++ {
			total, items, err := commentRepo.List(repository.CommentSubjectPost, repository.PostSubjectKey(post.ID), 1, offset)
			if err != nil {
				t.Fatal(err)
			}
			if total != 2 || len(items) != 1 {
				t.Fatalf("moderation changed pagination: total=%d items=%#v", total, items)
			}
			if offset == 0 && (items[0].ID != root.ID || items[0].Status != status || items[0].Content != root.Content) {
				t.Fatalf("moderated root or stored content changed: %#v", items[0])
			}
			if offset == 1 && (items[0].RootID == nil || *items[0].RootID != root.ID || items[0].Status != repository.StatusPublished) {
				t.Fatalf("reply relationship changed: %#v", items[0])
			}
		}
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

	if err := postRepo.SetStatus(post.ID, repository.StatusHidden); err != nil {
		t.Fatal(err)
	}
	if err := postRepo.SetStatus(post.ID, repository.StatusPublished); err != nil {
		t.Fatal(err)
	}
	if err := postRepo.SetCommentsLocked(post.ID, true); err != nil {
		t.Fatal(err)
	}
	pinOrder := int32(0)
	if err := postRepo.SetPinOrder(post.ID, &pinOrder); err != nil {
		t.Fatal(err)
	}
	moderated, err := postRepo.Find(post.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if !moderated.CommentsLocked || moderated.PinOrder == nil || *moderated.PinOrder != pinOrder {
		t.Fatalf("unexpected post subresources: %#v", moderated.Post)
	}
	if err := postRepo.SetCommentsLocked(post.ID, false); err != nil {
		t.Fatal(err)
	}
	if err := postRepo.SetPinOrder(post.ID, nil); err != nil {
		t.Fatal(err)
	}

	secondPost, err := postRepo.Create(repository.CreatePostInput{
		CategorySlug: category.Slug,
		Title:        "CaseSensitiveTitle", Content: "用于排序", AuthorID: 8, AuthorUsername: "bob",
		Attr: `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, newestPosts, err := postRepo.List(repository.PostFilter{Sort: repository.PostSortNewest}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(newestPosts) < 2 || newestPosts[0].ID != secondPost.ID {
		t.Fatalf("unexpected newest order: %#v", newestPosts)
	}
	_, viewedPosts, err := postRepo.List(repository.PostFilter{Sort: repository.PostSortViews}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(viewedPosts) < 2 || viewedPosts[0].ID != post.ID {
		t.Fatalf("unexpected views order: %#v", viewedPosts)
	}
	_, commentedPosts, err := postRepo.List(repository.PostFilter{Sort: repository.PostSortComments}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(commentedPosts) < 2 || commentedPosts[0].ID != post.ID {
		t.Fatalf("unexpected comments order: %#v", commentedPosts)
	}
	searchTotal, searchedPosts, err := postRepo.List(repository.PostFilter{Search: "casesensitivetitle"}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if searchTotal != 1 || len(searchedPosts) != 1 || searchedPosts[0].ID != secondPost.ID {
		t.Fatalf("unexpected case-insensitive search: total=%d items=%#v", searchTotal, searchedPosts)
	}
}
