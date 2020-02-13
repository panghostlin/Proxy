/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 04 February 2020 - 15:40:06
** @Filename:				Router.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 15:15:03
*******************************************************************************/

package			main

import			_ "os"
import			"github.com/microgolang/logs"
import			"github.com/valyala/fasthttp"
import			"github.com/buaazp/fasthttprouter"
import			"github.com/fasthttp/websocket"

var fastupgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

func	WithAuth(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		accessToken := ctx.Request.Header.Cookie(`accessToken`)

		if (accessToken == nil) {
			logs.Error(`No token`)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			return
		}

		isSuccess, memberID := CheckMemberCookie(ctx, string(accessToken))
		if (!isSuccess) {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			return
		}
		hashKey := ctx.Request.Header.Cookie(`hashKey`)

		ctx.SetUserValue(`memberID`, memberID)
		ctx.SetUserValue(`hashKey`, hashKey)
		h(ctx)
	})
}

func	InitRouter() func(*fasthttp.RequestCtx) {
	router := fasthttprouter.New()
	router.POST("/newMember/", CreateNewMember)
	router.POST("/loginMember/", LoginMember)
	router.POST("/checkMember/", WithAuth(CheckMember))

	router.POST("/uploadPicture/", WithAuth(UploadPicture))
	router.GET("/ws/uploadPicture/", WSUploadPicture)
	router.GET("/downloadPicture/:pictureSize/:pictureID", WithAuth(DownloadPicture))
	router.POST("/deletePictures/", WithAuth(DeletePictures))

	router.POST("/pictures/getby/member/", WithAuth(ListPicturesByMember))
	router.POST("/pictures/getby/album/", WithAuth(ListPicturesByAlbum))
	router.POST("/pictures/set/album/", WithAuth(SetPicturesAlbum))

	router.POST("/albums/create/", WithAuth(CreateAlbum))
	router.POST("/albums/list/", WithAuth(ListAlbums))
	router.POST("/albums/get/", WithAuth(GetAlbum))
	router.POST("/albums/delete/", WithAuth(DeleteAlbum))
	router.POST("/albums/set/cover/", WithAuth(SetAlbumCover))
	router.POST("/albums/set/name/", WithAuth(SetAlbumName))

	return router.Handler
}
