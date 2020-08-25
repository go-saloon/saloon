package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
	"github.com/soypat/curso/models"
)

func ReplyGet(c buffalo.Context) error {
	//reply := models.Reply{}
	//c.Set("reply",reply)
	return c.Render(200, r.HTML("replies/create.plush.html"))
}

func ReplyPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	reply := new(models.Reply)
	user := c.Value("current_user").(*models.User)
	if err := c.Bind(reply); err != nil {
		return errors.WithStack(err)
	}
	topic := c.Value("topic").(*models.Topic)
	topic.AddSubscriber(user.ID)
	reply.AuthorID = user.ID
	reply.Author = user
	reply.TopicID = topic.ID
	reply.Topic = topic

	verrs, err := tx.ValidateAndCreate(reply)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Update(topic); err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("reply", reply)
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("replies/create"))
	}
	c.Flash().Add("success", T.Translate(c, "reply-create-success"))

	//err = newReplyNotify(c, topic, reply)
	//if err != nil {
	//	return errors.WithStack(err)
	//}
	f := c.Value("forum").(*models.Forum)
	return c.Redirect(302, "topicGetPath()", render.Data{"forum_title":f.Title, "cat_title":c.Param("cat_title"),
		"tid":topic.ID})
}

func editReplyGet(c buffalo.Context) error {
	return c.Render(200,r.HTML("replies/create.plush.html"))
}
func editReplyPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	reply := new(models.Reply)
	if err := tx.Find(reply, c.Param("rid")); err != nil {
		return errors.WithStack(err)
	}
	if err := c.Bind(reply); err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Update(reply); err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", T.Translate(c,"reply-edit-success"))
	f := c.Value("forum").(*models.Forum)
	return c.Redirect(302, "topicGetPath()", render.Data{"forum_title":f.Title, "cat_title":c.Param("cat_title"),
		"tid":c.Param("tid")})
}

func SetCurrentReply(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		//reply := &models.Reply{}
		reply, err := loadReply(c,c.Param("rid"))
		if err != nil {
			c.Flash().Add("danger",T.Translate(c,"topic-not-found"))
			return c.Render(404,r.HTML("meta/404.plush.html"))
		}
		c.Logger().Infof("Reply got %s",reply)
		c.Set("reply",reply)
		return next(c)
	}
}

func DeleteReply(c buffalo.Context) error {
	reply, err := loadReply(c, c.Param("rid"))
	if err != nil {
		return errors.WithStack(err)
	}
	f := c.Value("forum").(*models.Forum)
	usr := c.Value("current_user").(*models.User)
	if !(usr.Role != "admin" ) && usr.ID != reply.AuthorID {
		c.Flash().Add("danger", "You are not authorized to delete this reply")
		return c.Redirect(302, "topicGetPath()", render.Data{"forum_title":f.Title, "cat_title":c.Param("cat_title"),
			"tid":c.Param("tid")})
	}
	tx := c.Value("tx").(*pop.Connection)
	reply.Deleted = true
	if err := tx.Update(reply); err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", "Reply deleted successfuly.")
	return c.Redirect(302, "topicGetPath()", render.Data{"forum_title":f.Title, "cat_title":c.Param("cat_title"),
		"tid":c.Param("tid")})

}

func loadReply(c buffalo.Context, id string) (*models.Reply, error) {
	tx := c.Value("tx").(*pop.Connection)
	reply := &models.Reply{}
	if err := c.Bind(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := tx.Find(reply, id); err != nil {
		return nil, c.Error(404, err)
	}
	topic := new(models.Topic)
	if err := tx.Find(topic, reply.TopicID); err != nil {
		return nil, c.Error(404, err)
	}
	usr := new(models.User)
	if err := tx.Find(usr, reply.AuthorID); err != nil {
		return nil, c.Error(404, err)
	}
	reply.Topic = topic
	reply.Author = usr
	return reply, nil
}