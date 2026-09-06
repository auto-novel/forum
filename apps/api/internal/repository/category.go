package repository

import (
	"auth/.gen/main/public/model"
	"auth/.gen/main/public/table"
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"
)

type Category = model.Category

type CategoryRepository interface {
	List() ([]Category, error)
	Find(id int64) (*Category, error)
	Create(slug string, bannerURL *string, attr string) (*Category, error)
	Update(id int64, slug string, bannerURL *string) (*Category, error)
}

type categoryRepository struct{ db *sql.DB }

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) List() ([]Category, error) {
	stmt := SELECT(table.Category.AllColumns).
		FROM(table.Category).
		ORDER_BY(table.Category.ID.ASC())
	var dest []Category
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func (r *categoryRepository) Find(id int64) (*Category, error) {
	stmt := SELECT(table.Category.AllColumns).
		FROM(table.Category).
		WHERE(table.Category.ID.EQ(Int64(id)))
	var dest Category
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *categoryRepository) Create(slug string, bannerURL *string, attr string) (*Category, error) {
	dest := Category{Slug: slug, BannerURL: bannerURL, Attr: attr}
	stmt := table.Category.INSERT(table.Category.Slug, table.Category.BannerURL, table.Category.Attr).
		MODEL(dest).
		RETURNING(table.Category.AllColumns)
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *categoryRepository) Update(id int64, slug string, bannerURL *string) (*Category, error) {
	stmt := table.Category.UPDATE(table.Category.Slug, table.Category.BannerURL).
		SET(String(slug), bannerURL).
		WHERE(table.Category.ID.EQ(Int64(id))).
		RETURNING(table.Category.AllColumns)
	var dest Category
	if err := stmt.Query(r.db, &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}
