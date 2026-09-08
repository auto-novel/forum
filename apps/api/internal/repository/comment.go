package repository

import (
	"auth/.gen/main/public/model"
	"auth/.gen/main/public/table"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
)

type Comment = model.Comment

const (
	CommentSubjectPost  int16 = 0
	CommentSubjectNovel int16 = 1
)

type CreateCommentInput struct {
	SubjectType    int16
	SubjectKey     string
	RootID         *int64
	Content        string
	AuthorID       int64
	AuthorUsername string
	Attr           string
}

type CommentRepository interface {
	List(subjectType int16, subjectKey string, limit, offset int64) (int64, []Comment, error)
	Find(subjectType int16, id int64) (*Comment, error)
	Create(input CreateCommentInput) (*Comment, error)
	Update(subjectType int16, id int64, content string) (*Comment, error)
	SetStatus(subjectType int16, id int64, status int16) error
}

type commentRepository struct{ db *sql.DB }

func NewCommentRepository(db *sql.DB) CommentRepository { return &commentRepository{db: db} }

func (r *commentRepository) List(subjectType int16, subjectKey string, limit, offset int64) (int64, []Comment, error) {
	condition := table.Comment.SubjectType.EQ(Int16(subjectType)).
		AND(table.Comment.SubjectKey.EQ(String(subjectKey))).
		AND(table.Comment.Status.EQ(Int16(StatusPublished)))
	countStmt := SELECT(COUNT(STAR)).FROM(table.Comment).WHERE(condition)
	var count struct{ Count int64 }
	if err := countStmt.Query(r.db, &count); err != nil {
		return 0, nil, err
	}
	stmt := SELECT(table.Comment.AllColumns).
		FROM(table.Comment).
		WHERE(condition).
		ORDER_BY(
			IntExp(COALESCE(table.Comment.RootID, table.Comment.ID)).ASC(),
			table.Comment.CreatedAt.ASC(),
			table.Comment.ID.ASC(),
		).
		LIMIT(limit).
		OFFSET(offset)
	var dest []Comment
	if err := stmt.Query(r.db, &dest); err != nil {
		return 0, nil, err
	}
	return count.Count, dest, nil
}

func (r *commentRepository) Find(subjectType int16, id int64) (*Comment, error) {
	stmt := SELECT(table.Comment.AllColumns).
		FROM(table.Comment).
		WHERE(table.Comment.ID.EQ(Int64(id)).
			AND(table.Comment.SubjectType.EQ(Int16(subjectType))))
	var dest Comment
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *commentRepository) Create(input CreateCommentInput) (*Comment, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var postID int64
	if input.SubjectType == CommentSubjectPost {
		postID, err = PostIDFromSubjectKey(input.SubjectKey)
		if err != nil {
			return nil, err
		}
		lockPost := SELECT(table.Post.AllColumns).
			FROM(table.Post).
			WHERE(table.Post.ID.EQ(Int64(postID)).AND(table.Post.Status.EQ(Int16(StatusPublished)))).
			FOR(UPDATE())
		var post Post
		if err := lockPost.Query(tx, &post); err != nil {
			return nil, err
		}
		if post.CommentsLocked {
			return nil, ErrCommentsLocked
		}
	}
	if input.RootID != nil {
		rootStmt := SELECT(table.Comment.AllColumns).
			FROM(table.Comment).
			WHERE(table.Comment.ID.EQ(Int64(*input.RootID)).
				AND(table.Comment.SubjectType.EQ(Int16(input.SubjectType))).
				AND(table.Comment.Status.EQ(Int16(StatusPublished))))
		var root Comment
		if err := rootStmt.Query(tx, &root); err != nil {
			return nil, err
		}
		if root.SubjectKey != input.SubjectKey || root.RootID != nil {
			return nil, fmt.Errorf("comment %d is not a root comment of subject %q", *input.RootID, input.SubjectKey)
		}
	}
	record := Comment{
		SubjectType:    input.SubjectType,
		SubjectKey:     input.SubjectKey,
		RootID:         input.RootID,
		Content:        input.Content,
		AuthorID:       input.AuthorID,
		AuthorUsername: input.AuthorUsername,
		Attr:           input.Attr,
	}
	insert := table.Comment.INSERT(
		table.Comment.SubjectType,
		table.Comment.SubjectKey,
		table.Comment.RootID,
		table.Comment.Content,
		table.Comment.AuthorID,
		table.Comment.AuthorUsername,
		table.Comment.Attr,
	).
		MODEL(record).
		RETURNING(table.Comment.AllColumns)
	if err := insert.Query(tx, &record); err != nil {
		return nil, err
	}
	if input.SubjectType == CommentSubjectPost {
		touchPost := table.Post.UPDATE(table.Post.CommentsCount, table.Post.ActiveAt).
			SET(table.Post.CommentsCount.ADD(Int32(1)), TimestampzT(time.Now())).
			WHERE(table.Post.ID.EQ(Int64(postID)))
		if _, err := touchPost.Exec(tx); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *commentRepository) Update(subjectType int16, id int64, content string) (*Comment, error) {
	stmt := table.Comment.UPDATE(table.Comment.Content, table.Comment.UpdatedAt).
		SET(String(content), TimestampzT(time.Now())).
		WHERE(table.Comment.ID.EQ(Int64(id)).
			AND(table.Comment.SubjectType.EQ(Int16(subjectType))).
			AND(table.Comment.Status.EQ(Int16(StatusPublished)))).
		RETURNING(table.Comment.AllColumns)
	var dest Comment
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *commentRepository) SetStatus(subjectType int16, id int64, status int16) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	lockComment := SELECT(table.Comment.AllColumns).
		FROM(table.Comment).
		WHERE(table.Comment.ID.EQ(Int64(id)).
			AND(table.Comment.SubjectType.EQ(Int16(subjectType)))).
		FOR(UPDATE())
	var record Comment
	if err := lockComment.Query(tx, &record); err != nil {
		return err
	}
	updateComment := table.Comment.UPDATE(table.Comment.Status, table.Comment.UpdatedAt).
		SET(Int16(status), TimestampzT(time.Now())).
		WHERE(table.Comment.ID.EQ(Int64(id)).
			AND(table.Comment.SubjectType.EQ(Int16(subjectType))))
	if _, err := updateComment.Exec(tx); err != nil {
		return err
	}
	if record.SubjectType != CommentSubjectPost ||
		record.Status == status ||
		(record.Status != StatusPublished && status != StatusPublished) {
		return tx.Commit()
	}
	postID, err := PostIDFromSubjectKey(record.SubjectKey)
	if err != nil {
		return err
	}
	if record.Status == StatusPublished {
		updatePost := table.Post.UPDATE(table.Post.CommentsCount).
			SET(IntExp(GREATEST(table.Post.CommentsCount.SUB(Int32(1)), Int32(0)))).
			WHERE(table.Post.ID.EQ(Int64(postID)))
		if _, err := updatePost.Exec(tx); err != nil {
			return err
		}
	} else if status == StatusPublished {
		updatePost := table.Post.UPDATE(table.Post.CommentsCount, table.Post.ActiveAt).
			SET(table.Post.CommentsCount.ADD(Int32(1)), TimestampzT(time.Now())).
			WHERE(table.Post.ID.EQ(Int64(postID)))
		if _, err := updatePost.Exec(tx); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func PostSubjectKey(postID int64) string {
	return strconv.FormatInt(postID, 10)
}

func PostIDFromSubjectKey(subjectKey string) (int64, error) {
	postID, err := strconv.ParseInt(subjectKey, 10, 64)
	if err != nil || postID <= 0 {
		return 0, fmt.Errorf("invalid post subject key %q", subjectKey)
	}
	return postID, nil
}
