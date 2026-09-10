package repository

import (
	"auth/.gen/main/public/model"
	"auth/.gen/main/public/table"
	"database/sql"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

type Post = model.Post

type PostDetails struct {
	model.Post
	Tags []Tag
}

const (
	PostSortActive   = "active"
	PostSortNewest   = "newest"
	PostSortViews    = "views"
	PostSortComments = "comments"
)

type PostFilter struct {
	CategorySlug, Search     string
	Sort                     string
	TagIDs                   []int64
	AuthorID, FavoriteUserID int64
}

type CreatePostInput struct {
	CategorySlug, Title, Content string
	AuthorID                     int64
	AuthorUsername               string
	TagIDs                       []int64
	Attr                         string
}

type UpdatePostInput struct {
	Title, Content string
	TagIDs         []int64
}

type PostRepository interface {
	List(filter PostFilter, limit, offset int64) (int64, []PostDetails, error)
	Find(id int64, incrementViews bool) (*PostDetails, error)
	Create(input CreatePostInput) (*PostDetails, error)
	Update(id int64, input UpdatePostInput) (*PostDetails, error)
	SetStatus(id int64, status int16) error
	SetModeration(id int64, status int16, commentsLocked bool, pinOrder *int32) error
}

type postRepository struct {
	db      *sql.DB
	tagRepo TagRepository
}

func NewPostRepository(db *sql.DB, tagRepo TagRepository) PostRepository {
	return &postRepository{db: db, tagRepo: tagRepo}
}

func integerExpressions(ids []int64) []Expression {
	expressions := make([]Expression, len(ids))
	for i, id := range ids {
		expressions[i] = Int64(id)
	}
	return expressions
}

func (filter PostFilter) condition() BoolExpression {
	expressions := []BoolExpression{table.Post.Status.EQ(Int16(StatusPublished))}
	if filter.CategorySlug != "" {
		expressions = append(expressions, table.Category.Slug.EQ(String(filter.CategorySlug)))
	}
	if filter.AuthorID > 0 {
		expressions = append(expressions, table.Post.AuthorID.EQ(Int64(filter.AuthorID)))
	}
	if filter.FavoriteUserID > 0 {
		expressions = append(expressions, table.PostFavorite.UserID.EQ(Int64(filter.FavoriteUserID)))
	}
	if filter.Search != "" {
		pattern := String("%" + filter.Search + "%")
		expressions = append(expressions, OR(table.Post.Title.LIKE(pattern), table.Post.Content.LIKE(pattern)))
	}
	if len(filter.TagIDs) > 0 {
		taggedPosts := SELECT(table.PostTag.PostID).
			FROM(table.PostTag).
			WHERE(table.PostTag.TagID.IN(integerExpressions(filter.TagIDs)...)).
			GROUP_BY(table.PostTag.PostID).
			HAVING(COUNT(table.PostTag.TagID).EQ(Int64(int64(len(filter.TagIDs)))))
		expressions = append(expressions, table.Post.ID.IN(taggedPosts))
	}
	return AND(expressions...)
}

func postFrom(filter PostFilter) ReadableTable {
	from := table.Post.INNER_JOIN(table.Category, table.Post.CategoryID.EQ(table.Category.ID))
	if filter.FavoriteUserID > 0 {
		return from.INNER_JOIN(table.PostFavorite, table.PostFavorite.PostID.EQ(table.Post.ID))
	}
	return from
}

func postOrderBy(sort string) []OrderByClause {
	pinned := table.Post.PinOrder.ASC().NULLS_LAST()
	switch sort {
	case PostSortNewest:
		return []OrderByClause{pinned, table.Post.CreatedAt.DESC(), table.Post.ID.DESC()}
	case PostSortViews:
		return []OrderByClause{pinned, table.Post.ViewsCount.DESC(), table.Post.ActiveAt.DESC(), table.Post.ID.DESC()}
	case PostSortComments:
		return []OrderByClause{pinned, table.Post.CommentsCount.DESC(), table.Post.ActiveAt.DESC(), table.Post.ID.DESC()}
	default:
		return []OrderByClause{pinned, table.Post.ActiveAt.DESC(), table.Post.ID.DESC()}
	}
}

func (r *postRepository) List(filter PostFilter, limit, offset int64) (int64, []PostDetails, error) {
	condition := filter.condition()
	from := postFrom(filter)
	countStmt := SELECT(COUNT(STAR)).FROM(from).WHERE(condition)
	var count struct{ Count int64 }
	if err := countStmt.Query(r.db, &count); err != nil {
		return 0, nil, err
	}

	stmt := SELECT(table.Post.AllColumns).
		FROM(from).
		WHERE(condition).
		ORDER_BY(postOrderBy(filter.Sort)...).
		LIMIT(limit).
		OFFSET(offset)
	var records []Post
	if err := stmt.Query(r.db, &records); err != nil {
		return 0, nil, err
	}
	dest := make([]PostDetails, len(records))
	for i, record := range records {
		tags, err := r.tagRepo.ListForPost(record.ID)
		if err != nil {
			return 0, nil, err
		}
		dest[i] = PostDetails{Post: record, Tags: tags}
	}
	return count.Count, dest, nil
}

func (r *postRepository) Find(id int64, incrementViews bool) (*PostDetails, error) {
	var dest Post
	var err error
	if incrementViews {
		stmt := table.Post.UPDATE(table.Post.ViewsCount).
			SET(table.Post.ViewsCount.ADD(Int32(1))).
			WHERE(table.Post.ID.EQ(Int64(id)).AND(table.Post.Status.EQ(Int16(StatusPublished)))).
			RETURNING(table.Post.AllColumns)
		err = stmt.Query(r.db, &dest)
	} else {
		stmt := SELECT(table.Post.AllColumns).
			FROM(table.Post).
			WHERE(table.Post.ID.EQ(Int64(id)).AND(table.Post.Status.EQ(Int16(StatusPublished))))
		err = stmt.Query(r.db, &dest)
	}
	if err != nil {
		return nil, err
	}
	tags, err := r.tagRepo.ListForPost(id)
	if err != nil {
		return nil, err
	}
	return &PostDetails{Post: dest, Tags: tags}, nil
}

func (r *postRepository) Create(input CreatePostInput) (*PostDetails, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var category Category
	findCategory := SELECT(table.Category.AllColumns).
		FROM(table.Category).
		WHERE(table.Category.Slug.EQ(String(input.CategorySlug)))
	if err := findCategory.Query(tx, &category); err != nil {
		return nil, err
	}
	record := Post{
		CategoryID:     category.ID,
		Title:          input.Title,
		AuthorID:       input.AuthorID,
		AuthorUsername: input.AuthorUsername,
		Content:        input.Content,
		Attr:           input.Attr,
	}
	insert := table.Post.INSERT(
		table.Post.CategoryID,
		table.Post.Title,
		table.Post.AuthorID,
		table.Post.AuthorUsername,
		table.Post.Content,
		table.Post.Attr,
	).
		MODEL(record).
		RETURNING(table.Post.AllColumns)
	if err := insert.Query(tx, &record); err != nil {
		return nil, err
	}
	if err := replacePostTags(tx, record.ID, record.CategoryID, input.TagIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tags, err := r.tagRepo.ListForPost(record.ID)
	if err != nil {
		return nil, err
	}
	return &PostDetails{Post: record, Tags: tags}, nil
}

func replacePostTags(db qrm.DB, postID, categoryID int64, tagIDs []int64) error {
	deleteStmt := table.PostTag.DELETE().WHERE(table.PostTag.PostID.EQ(Int64(postID)))
	if _, err := deleteStmt.Exec(db); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	validStmt := SELECT(table.Tag.AllColumns).
		FROM(table.Tag).
		WHERE(table.Tag.CategoryID.EQ(Int64(categoryID)).
			AND(table.Tag.IsActive.IS_TRUE()).
			AND(table.Tag.ID.IN(integerExpressions(tagIDs)...)))
	var valid []Tag
	if err := validStmt.Query(db, &valid); err != nil {
		return err
	}
	if len(valid) != len(tagIDs) {
		return ErrInvalidTag
	}
	links := make([]model.PostTag, len(valid))
	for i, tag := range valid {
		links[i] = model.PostTag{PostID: postID, TagID: tag.ID}
	}
	insertStmt := table.PostTag.INSERT(table.PostTag.PostID, table.PostTag.TagID).MODELS(links)
	_, err := insertStmt.Exec(db)
	return err
}

func (r *postRepository) Update(id int64, input UpdatePostInput) (*PostDetails, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	stmt := table.Post.UPDATE(table.Post.Title, table.Post.Content, table.Post.UpdatedAt).
		SET(String(input.Title), String(input.Content), TimestampzT(time.Now())).
		WHERE(table.Post.ID.EQ(Int64(id)).AND(table.Post.Status.EQ(Int16(StatusPublished)))).
		RETURNING(table.Post.AllColumns)
	var record Post
	if err := stmt.Query(tx, &record); err != nil {
		return nil, err
	}
	if err := replacePostTags(tx, id, record.CategoryID, input.TagIDs); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tags, err := r.tagRepo.ListForPost(id)
	if err != nil {
		return nil, err
	}
	return &PostDetails{Post: record, Tags: tags}, nil
}

func (r *postRepository) SetStatus(id int64, status int16) error {
	stmt := table.Post.UPDATE(table.Post.Status, table.Post.UpdatedAt).
		SET(Int16(status), TimestampzT(time.Now())).
		WHERE(table.Post.ID.EQ(Int64(id)).AND(table.Post.Status.NOT_EQ(Int16(StatusDeleted))))
	result, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return qrm.ErrNoRows
	}
	return nil
}

func (r *postRepository) SetModeration(id int64, status int16, commentsLocked bool, pinOrder *int32) error {
	stmt := table.Post.UPDATE(table.Post.Status, table.Post.CommentsLocked, table.Post.PinOrder, table.Post.UpdatedAt).
		SET(Int16(status), Bool(commentsLocked), pinOrder, TimestampzT(time.Now())).
		WHERE(table.Post.ID.EQ(Int64(id)))
	result, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return qrm.ErrNoRows
	}
	return nil
}
