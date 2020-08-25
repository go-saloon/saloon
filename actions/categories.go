package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
	"github.com/soypat/curso/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// CategoriesIndex default implementation.
func CategoriesIndex(c buffalo.Context) error {
	catTitle := c.Param("cat_title")
	c.Logger().Printf("accessed %s",catTitle)
	tx := c.Value("tx").(*pop.Connection)
	cat := &models.Category{}

	err:=tx.Where("title = ?",catTitle).First(cat)
	if  err != nil {
		return c.Error(404, err)
	}
	c.Set("category", cat)
	topics := &models.Topics{}
	if err := tx.BelongsTo(cat).All(topics); err != nil {
		return c.Error(404, err)
	}
	c.Set("topics", topics)
	for i, t := range *topics {
		topic, err := loadTopic(c, t.ID.String())
		if err != nil {
			return errors.WithStack(err)
		}
		(*topics)[i] = *topic
	}
	return c.Render(200, r.HTML("categories/index.plush.html"))
	//return c.Render(http.StatusOK, r.HTML("categories/index.plush.html"))
}

// CategoriesCreateGet default implementation.
func CategoriesCreateGet(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("categories/create_get.plush.html"))
}

// CategoriesCreatePost default implementation.
func CategoriesCreatePost(c buffalo.Context) error {
	cat := &models.Category{}
	if err := c.Bind(cat); err != nil {
		c.Flash().Add("danger", "could not create forum")
		c.Logger().Printf("error: %s", err)
		return c.Redirect(302, "")
	}
	c.Logger().Printf("T %s %s", cat.Title, cat.Description)
	if !validURLDir(cat.Title) {

		c.Flash().Add("danger", T.Translate(c,"category-invalid-title"))
		return c.Redirect(302, "")
	}
	f := c.Value("forum").(*models.Forum)
	cat.ParentCategory = nulls.NewUUID(f.ID)
	tx := c.Value("tx").(*pop.Connection)
	q := tx.Where("title = ?", cat.Title)
	exist, err := q.Exists(&models.Forum{})
	if exist {
		c.Flash().Add("danger", "Category already exists")
		//return c.Render(200,r.HTML("forums/create.plush.html"))
		return c.Redirect(302, "")
	}
	v, _ := cat.Validate(tx)
	if v.HasAny() {
		c.Flash().Add("danger", "Title should have something!")
		return c.Redirect(302, "")
	}
	err = tx.Save(cat)
	if err != nil {
		c.Flash().Add("danger", "Error creating category")
		return errors.WithStack(err)
	}
	c.Flash().Add("success", fmt.Sprintf("Category %s created", cat.Title))
	return c.Redirect(302, "forumPath()",render.Data{"forum_title":f.Title})
}

// CategoriesDetail default implementation.
func CategoriesDetail(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("categories/detail.html"))
}

// SetCurrentCategory attempts to find a category and set context `category`
func SetCurrentCategory(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		tx := c.Value("tx").(*pop.Connection)
		cat := &models.Category{}
		title := c.Param("cat_title")
		if title == "" {
			return c.Redirect(302,"forumPath()")
		}
		q  := tx.Where("title = ?", title)
		err := q.First(cat)
		if err != nil {
			c.Flash().Add("danger","Error while seeking category")
			return c.Redirect(302,"forumPath()")
		}
		c.Set("inCat",true)
		c.Set("category", cat)
		return next(c)
	}
}
