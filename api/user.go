package api

import (
	"github.com/Nikita-Boyarskikh/DB/db"
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/opt"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func UserRouter(user *routing.RouteGroup) {
	user.Post("/<nickname>/create", func(ctx *routing.Context) error {
		var (
			nickname = ctx.Param("nickname")
			user     = models.User{
				Nickname: opt.OString(nickname),
			}
		)

		if err := easyjson.Unmarshal(ctx.PostBody(), &user); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		users, err := db.GetUsersByEmailAndNickname(user.Email, nickname)
		if len(users) > 0 {
			json, err := easyjson.Marshal(users)
			if err != nil {
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(json)
			return nil
		} else if err != nil {
			return err
		}

		if err := db.CreateUser(user); err != nil {
			return err
		}

		json, err := easyjson.Marshal(user)
		if err != nil {
			return err
		}

		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		ctx.SetBody(json)
		return nil
	})

	user.Post("/<nickname>/profile", func(ctx *routing.Context) error {
		var (
			nickname = ctx.Param("nickname")
			user     = models.User{
				Nickname: opt.OString(nickname),
			}
		)

		if err := easyjson.Unmarshal(ctx.PostBody(), &user); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.WriteData(err.Error())
			return nil
		}

		users, err := db.GetUsersByEmailAndNickname(user.Email, nickname)
		if len(users) > 1 {
			message, err := easyjson.Marshal(models.Error{"Email or nickname conflict with existing users"})
			if err != nil {
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetContentType("application/json")
			ctx.SetBody(message)
			return nil
		} else if len(users) == 0 {
			message, err := easyjson.Marshal(models.Error{"User with requested nickname is not found"})
			if err != nil {
				return err
			}

			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetContentType("application/json")
			ctx.SetBody(message)

			return nil
		} else if err != nil {
			return err
		}

		updatedUser, err := db.UpdateUser(user)
		if err != nil {
			return err
		}

		json, err := easyjson.Marshal(updatedUser)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)

		return nil
	})

	user.Get("/<nickname>/profile", func(ctx *routing.Context) error {
		nickname := ctx.Param("nickname")

		user, err := db.GetUserByNickname(nickname)
		if err != nil {
			if err == pgx.ErrNoRows {
				json, err := easyjson.Marshal(models.Error{"Requested user is not found"})
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

		json, err := easyjson.Marshal(user)
		if err != nil {
			return err
		}

		ctx.Success("application/json", json)
		return nil
	})
}
