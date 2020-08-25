package actions

import (
	"fmt"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
	"github.com/soypat/curso/models"
	"regexp"
	"strings"
)

func UserSettingsGet(c buffalo.Context) error {
	return c.Render(200, r.HTML("users/settings.plush.html"))
}

func UserSettingsPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	userDB := c.Value("current_user").(*models.User)
	user := &models.User{}
	if err := c.Bind(user); err != nil {
		return errors.WithStack(err)
	}
	s, err := sanitizeNick(user.Nick)
	if err != nil {
		c.Flash().Add("danger",T.Translate(c,"user-settings-nick-invalid"))
		return c.Redirect(302, "userSettingsPath()")
	}
	user.Nick = s
	if userDB.Nick == "" && user.Nick == "" {
		c.Flash().Add("danger",T.Translate(c, "user-settings-nick-empty"))
		return c.Redirect(302, "userSettingsPath()")
	}
	userDB.Nick = user.Nick
	if err := tx.Update(userDB); err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", T.Translate(c,"user-settings-nick-edit-success"))
	return c.Redirect(302, "userSettingsPath()") //return c.Redirect(302, "topicGetPath()", render.Data{"forum_title":f.Title, "cat_title":c.Param("cat_title"),
		// "tid":c.Param("tid")})
}

func sanitizeNick(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > 10 {
		return "", errors.New("nick too long")
	}
	re := regexp.MustCompile(fmt.Sprintf("[0-9a-zA-Z]{%d}",len(s)))
	if !re.MatchString(s) {
		return s,errors.New("nick contains non alphanumeric char")
	}
	return s, nil
}