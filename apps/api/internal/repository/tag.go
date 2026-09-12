package repository

import (
	"auth/.gen/main/public/model"
	"auth/.gen/main/public/table"
	"database/sql"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

type Tag = model.Tag

type TagRepository interface {
	ListByCategory(categoryID int64) ([]Tag, error)
	ListActive() ([]Tag, error)
	ListForPost(postID int64) ([]Tag, error)
	Create(categoryID int64, name string, color int16, sortOrder int32, attr string) (*Tag, error)
	Update(id int64, name string, color int16, sortOrder int32) (*Tag, error)
	SetActive(id int64, active bool) error
}

type tagRepository struct{ db *sql.DB }

func NewTagRepository(db *sql.DB) TagRepository { return &tagRepository{db: db} }

func (r *tagRepository) ListByCategory(categoryID int64) ([]Tag, error) {
	stmt := SELECT(table.Tag.AllColumns).
		FROM(table.Tag).
		WHERE(table.Tag.CategoryID.EQ(Int64(categoryID))).
		ORDER_BY(table.Tag.SortOrder.ASC(), table.Tag.ID.ASC())
	var dest []Tag
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func (r *tagRepository) ListActive() ([]Tag, error) {
	stmt := SELECT(table.Tag.AllColumns).
		FROM(table.Tag).
		WHERE(table.Tag.IsActive.IS_TRUE()).
		ORDER_BY(table.Tag.CategoryID.ASC(), table.Tag.SortOrder.ASC(), table.Tag.ID.ASC())
	var dest []Tag
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func (r *tagRepository) ListForPost(postID int64) ([]Tag, error) {
	stmt := SELECT(table.Tag.AllColumns).
		FROM(table.Tag.INNER_JOIN(table.PostTag, table.Tag.ID.EQ(table.PostTag.TagID))).
		WHERE(table.PostTag.PostID.EQ(Int64(postID))).
		ORDER_BY(table.Tag.SortOrder.ASC(), table.Tag.ID.ASC())
	var dest []Tag
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func (r *tagRepository) Create(categoryID int64, name string, color int16, sortOrder int32, attr string) (*Tag, error) {
	dest := Tag{CategoryID: categoryID, Name: name, Color: color, SortOrder: sortOrder, Attr: attr}
	stmt := table.Tag.INSERT(table.Tag.CategoryID, table.Tag.Name, table.Tag.Color, table.Tag.SortOrder, table.Tag.Attr).
		MODEL(dest).
		RETURNING(table.Tag.AllColumns)
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *tagRepository) Update(id int64, name string, color int16, sortOrder int32) (*Tag, error) {
	stmt := table.Tag.UPDATE(
		table.Tag.Name,
		table.Tag.Color,
		table.Tag.SortOrder,
		table.Tag.UpdatedAt,
	).
		SET(String(name), Int16(color), Int32(sortOrder), TimestampzT(time.Now())).
		WHERE(table.Tag.ID.EQ(Int64(id))).
		RETURNING(table.Tag.AllColumns)
	var dest Tag
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *tagRepository) SetActive(id int64, active bool) error {
	stmt := table.Tag.UPDATE(table.Tag.IsActive, table.Tag.UpdatedAt).
		SET(Bool(active), TimestampzT(time.Now())).
		WHERE(table.Tag.ID.EQ(Int64(id)))
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
