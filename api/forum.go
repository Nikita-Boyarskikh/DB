package api

import (
	"time"

	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/opt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func ForumRouter(forum *routing.RouteGroup) {
	forum.Post("/create", func(ctx *routing.Context) error {
		var forum models.Forum
		if err := easyjson.Unmarshal(ctx.PostBody(), &forum); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return nil
		}

		if _, exForum, err := db.GetForumBySlug(forum.Slug); err == nil {
			json, err := easyjson.Marshal(exForum)
			if err != nil {
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(json)

			return nil
		} else if err != pgx.ErrNoRows {
			return err
		}

		user, err := db.GetUserByNickname(forum.User)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Forums author is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			} else {
				return err
			}
		}

		forum.User = user.Nickname.V
		if _, err := db.CreateForum(forum); err != nil {
			return err
		}

		json, err := easyjson.Marshal(forum)
		if err != nil {
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)

		db.NewForum()
		return nil
	})

	forum.Get("/<slug>/details", func(ctx *routing.Context) error {
		slug := ctx.Param("slug")

		_, forum, err := db.GetForumBySlug(slug)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Forum with requested slug is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			} else {
				return err
			}
		}

		json, err := easyjson.Marshal(forum)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})

	forum.Post("/<slug>/create", func(ctx *routing.Context) error {
		slug := ctx.Param("slug")

		_, forum, err := db.GetForumBySlug(slug)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Forum with requested slug is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			} else {
				return err
			}
		}

		var thread models.Thread
		if err := easyjson.Unmarshal(ctx.PostBody(), &thread); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return nil
		}

		user, err := db.GetUserByNickname(thread.Author)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"User who thread's author is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			} else {
				return err
			}
		}

		if exTread, err := db.GetThreadBySlugOrID(thread.Slug.V, thread.ID.V); err == nil {
			json, err := easyjson.Marshal(exTread)
			if err != nil {
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(json)

			return nil
		} else if err != pgx.ErrNoRows {
			return err
		}

		thread.Author = user.Nickname.V
		thread.Forum = opt.OString(forum.Slug)
		if thread, err = db.CreateThread(thread); err != nil {
			return err
		}

		json, err := easyjson.Marshal(thread)
		if err != nil {
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)

		return nil
	})

	forum.Get("/<slug>/threads", func(ctx *routing.Context) error {
		slug := ctx.Param("slug")

		_, _, err := db.GetForumBySlug(slug)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Forum with requested slug is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			} else {
				return err
			}
		}

		limit, err := ctx.QueryArgs().GetUint("limit")
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return nil
		}
		sinceRaw := ctx.QueryArgs().Peek("since")
		descRaw := ctx.QueryArgs().Peek("desc")
		desc := string(descRaw) == "true"

		var sinceTime time.Time
		if len(sinceRaw) > 0 {
			if sinceTime, err = time.Parse(config.TimestampInLayout, string(sinceRaw)); err != nil {
				ctx.SetStatusCode(fasthttp.StatusBadRequest)
				return nil
			}
		} else {
			if desc {
				sinceTime = time.Now().AddDate(9999, 0, 0)
			} else {
				sinceTime = time.Unix(0, 0)
			}
		}

		threads, err := db.GetThreadsBySlug(slug, limit, sinceTime, desc)
		if err != nil {
			return err
		}

		if len(threads) == 0 {
			threads = models.Threads{}
		}

		json, err := easyjson.Marshal(threads)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})

	forum.Get("/<slug>/users", func(ctx *routing.Context) error {
		slug := ctx.Param("slug")

		_, _, err := db.GetForumBySlug(slug)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Forum with requested slug is not found"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			} else {
				return err
			}
		}

		limit, err := ctx.QueryArgs().GetUint("limit")
		if err != nil {
			limit = -1
		}
		since := ctx.QueryArgs().Peek("since")
		descRaw := ctx.QueryArgs().Peek("desc")
		desc := string(descRaw) == "true"

		users, err := db.GetUsersByForumSlug(slug, string(since), limit, desc)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			users = models.Users{}
		}

		json, err := easyjson.Marshal(users)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})
}
