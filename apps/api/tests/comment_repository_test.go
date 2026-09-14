//go:build integration

package tests

import (
	forumcategory "auth/internal/category"
	"auth/internal/repository"
	"testing"
)

func TestCommentRepositoryDeleteAllByAuthor(t *testing.T) {
	resetDatabase()
	category, _ := forumcategory.FindByID(forumcategory.NovelID)
	firstPost, err := postRepo.Create(repository.CreatePostInput{
		CategoryID: category.ID, Title: "第一篇帖子", Content: "正文",
		AuthorID: 1, AuthorUsername: "author", Attr: `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	secondPost, err := postRepo.Create(repository.CreatePostInput{
		CategoryID: category.ID, Title: "第二篇帖子", Content: "正文",
		AuthorID: 1, AuthorUsername: "author", Attr: `{}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	createComment := func(subjectType int16, subjectKey string, authorID int64) *repository.Comment {
		t.Helper()
		comment, err := commentRepo.Create(repository.CreateCommentInput{
			SubjectType: subjectType, SubjectKey: subjectKey, Content: "评论",
			AuthorID: authorID, AuthorUsername: "commenter", Attr: `{}`,
		})
		if err != nil {
			t.Fatal(err)
		}
		return comment
	}

	firstPublished := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(firstPost.ID), 7)
	firstHidden := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(firstPost.ID), 7)
	if err := commentRepo.SetStatus(repository.CommentSubjectPost, firstHidden.ID, repository.StatusHidden); err != nil {
		t.Fatal(err)
	}
	secondPublished := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(secondPost.ID), 7)
	external := createComment(repository.CommentSubjectNovel, "novel:chapter-1", 7)
	otherAuthor := createComment(repository.CommentSubjectPost, repository.PostSubjectKey(firstPost.ID), 8)

	if err := commentRepo.DeleteAllByAuthor(7); err != nil {
		t.Fatal(err)
	}
	// 重复删除必须保持幂等，不能再次扣减帖子评论数。
	if err := commentRepo.DeleteAllByAuthor(7); err != nil {
		t.Fatal(err)
	}

	for subjectType, commentIDs := range map[int16][]int64{
		repository.CommentSubjectPost:  {firstPublished.ID, firstHidden.ID, secondPublished.ID},
		repository.CommentSubjectNovel: {external.ID},
	} {
		for _, commentID := range commentIDs {
			comment, err := commentRepo.Find(subjectType, commentID)
			if err != nil {
				t.Fatal(err)
			}
			if comment.Status != repository.StatusDeleted {
				t.Fatalf("comment %d status = %d", comment.ID, comment.Status)
			}
		}
	}

	remaining, err := commentRepo.Find(repository.CommentSubjectPost, otherAuthor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if remaining.Status != repository.StatusPublished {
		t.Fatalf("other author's comment status = %d", remaining.Status)
	}

	first, err := postRepo.Find(firstPost.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := postRepo.Find(secondPost.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if first.CommentsCount != 1 || second.CommentsCount != 0 {
		t.Fatalf("unexpected comment counts: first=%d second=%d", first.CommentsCount, second.CommentsCount)
	}
}
