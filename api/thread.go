package api

import (
	"strconv"

	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/opt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func ThreadRouter(thread *routing.RouteGroup) {
	thread.Get("/<slug_or_id>/details", func(ctx *routing.Context) error {
		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		if thread, err := db.GetThreadBySlugOrID(slug_or_id, int32(id)); err == nil {
			json, err := easyjson.Marshal(thread)
			if err != nil {
				return err
			}

			ctx.Success("application/json", json)
			return nil
		} else if err != pgx.ErrNoRows {
			return err
		}

		json, err := easyjson.Marshal(models.Error{"Thread with requested slug or id is not found"})
		if err != nil {
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)
		return nil
	})

	thread.Post("/<slug_or_id>/details", func(ctx *routing.Context) error {
		slug_or_id := ctx.Param("slug_or_id")

		id, _ := strconv.Atoi(slug_or_id)

		if _, err := db.GetThreadBySlugOrID(slug_or_id, int32(id)); err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Thread with requested slug or ID is not found"})
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

		var patch models.PatchThread
		if err := easyjson.Unmarshal(ctx.PostBody(), &patch); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		thread, err := db.UpdateThreadBySlugOrID(slug_or_id, patch)
		if err != nil {
			return err
		}

		json, err := easyjson.Marshal(thread)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})

	thread.Post("/<slug_or_id>/vote", func(ctx *routing.Context) error {
		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		t, err := db.GetThreadBySlugOrID(slug_or_id, int32(id))
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Thread with requested slug or ID is not found"})
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

		var vote models.Vote
		if err := easyjson.Unmarshal(ctx.PostBody(), &vote); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		_, err = db.GetUserByNickname(vote.Nickname)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"User with requested nickname is not found"})
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

		thread, err := db.VoteForThread(&t, vote)
		if err != nil {
			return err
		}

		json, err := easyjson.Marshal(*thread)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})

	thread.Post("/<slug_or_id>/create", func(ctx *routing.Context) error {
		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		t, err := db.GetThreadBySlugOrID(slug_or_id, int32(id))
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Thread with requested slug or ID is not found"})
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

		var posts models.Posts
		if err := easyjson.Unmarshal(ctx.PostBody(), &posts); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		var (
			nicknames           []string
			parents             []int64
			parentsWithoutZeros []string
		)
		for _, post := range posts {
			nicknames = append(nicknames, post.Author)
			if !post.Parent.Defined {
				post.Parent = opt.OInt64(0)
			} else {
				parentsWithoutZeros = append(parentsWithoutZeros, strconv.FormatInt(post.Parent.V, 10))
			}
			parents = append(parents, post.Parent.V)
		}

		if len(nicknames) > 0 {
			exists, _ := db.CheckAllUsersExists(nicknames)
			if !exists {
				json, err := easyjson.Marshal(models.Error{"Can't find any post authors"})
				if err != nil {
					return err
				}

				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetContentType("application/json")
				ctx.SetBody(json)
				return nil
			}
		}

		var otherThread bool
		if otherThread, err = db.CheckAllPostsInOneThread(t.ID.V, posts); err != nil {
			return err
		} else if !otherThread {
			json, err := easyjson.Marshal(models.Error{"Parent post was created in another thread"})
			if err != nil {
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(json)
			return nil
		}

		createdPosts, err := db.CreatePostsInThread(t.Forum.V, t.ID.V, posts, parents, parentsWithoutZeros)
		if err != nil {
			return err
		}

		json, err := easyjson.Marshal(createdPosts)
		if err != nil {
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)
		return nil
	})

	thread.Get("/<slug_or_id>/posts", func(ctx *routing.Context) error {
		slug_or_id := ctx.Param("slug_or_id")
		id, _ := strconv.Atoi(slug_or_id)

		t, err := db.GetThreadBySlugOrID(slug_or_id, int32(id))
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Thread with requested slug or ID is not found"})
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
		sinceRow := ctx.QueryArgs().Peek("since")
		descRaw := ctx.QueryArgs().Peek("desc")
		sort := ctx.QueryArgs().Peek("sort")
		desc := string(descRaw) == "true"

		var since int64
		sinceInt, err := strconv.Atoi(string(sinceRow))
		if err != nil {
			since = -1
		} else {
			since = int64(sinceInt)
		}

		posts, err := db.GetPosts(t.ID.V, limit, since, string(sort), desc)
		if err != nil {
			return err
		}

		if len(posts) == 0 {
			posts = models.Posts{}
		}

		json, err := easyjson.Marshal(posts)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})
}
