/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 14 January 2020 - 20:21:56
** @Filename:				Albums.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Wednesday 01 April 2020 - 12:06:16
*******************************************************************************/

package			main

import			"context"
import			"github.com/microgolang/logs"
import			"github.com/panghostlin/SDK/Pictures"
import			"github.com/valyala/fasthttp"
import			"encoding/json"

/******************************************************************************
**	downloadPictureGRPC
**	Call the Picture Microservice to download an image.
**
**	DownloadPicture
**	Router proxy function to download an image.
******************************************************************************/
func	createAlbumGRPC(memberID string, req *pictures.CreateAlbumRequest) (*pictures.CreateAlbumResponse, error) {
	result, err := clients.albums.CreateAlbum(
		context.Background(),
		&pictures.CreateAlbumRequest{
			Name: req.GetName(),
			MemberID: memberID,
			CoverPicture: req.GetCoverPicture(),
			Pictures: req.GetPictures(),
		},
	)

	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	createAlbum(ctx *fasthttp.RequestCtx) {
	req := &pictures.CreateAlbumRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	memberID := ctx.UserValue("memberID").(string)

	data, err := createAlbumGRPC(memberID, req)
	resolve(ctx, data, err, 401)
}

/******************************************************************************
**	listAlbumsGRPC
******************************************************************************/
func	listAlbumsGRPC(memberID string) (*pictures.ListAlbumsResponse, error) {
	result, err := clients.albums.ListAlbums(
		context.Background(),
		&pictures.ListAlbumsRequest{MemberID: memberID},
	)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	listAlbums(ctx *fasthttp.RequestCtx) {
	memberID := ctx.UserValue("memberID").(string)
	data, err := listAlbumsGRPC(memberID)

	resolve(ctx, data.GetAlbums(), err, 401)
}


/******************************************************************************
**	setAlbumCoverGRPC
******************************************************************************/
func	setAlbumCoverGRPC(req *pictures.SetAlbumCoverRequest) (*pictures.SetAlbumCoverResponse, error) {
	result, err := clients.albums.SetAlbumCover(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	setAlbumCover(ctx *fasthttp.RequestCtx) {
	req := &pictures.SetAlbumCoverRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := setAlbumCoverGRPC(req)
	resolve(ctx, data.GetAlbumID(), err, 401)
}


/******************************************************************************
**	setAlbumNameGRPC
******************************************************************************/
func	setAlbumNameGRPC(req *pictures.SetAlbumNameRequest) (*pictures.SetAlbumNameResponse, error) {
	result, err := clients.albums.SetAlbumName(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	setAlbumName(ctx *fasthttp.RequestCtx) {
	req := &pictures.SetAlbumNameRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := setAlbumNameGRPC(req)
	resolve(ctx, data.GetAlbumID(), err, 401)
}


/******************************************************************************
**	deleteAlbumGRPC
******************************************************************************/
func	deleteAlbumGRPC(req *pictures.DeleteAlbumRequest) (*pictures.DeleteAlbumResponse, error) {
	result, err := clients.albums.DeleteAlbum(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	deleteAlbum(ctx *fasthttp.RequestCtx) {
	req := &pictures.DeleteAlbumRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := deleteAlbumGRPC(req)
	resolve(ctx, data.GetSuccess(), err, 401)
}


/******************************************************************************
**	getAlbumGRPC
******************************************************************************/
func	getAlbumGRPC(req *pictures.GetAlbumRequest) (*pictures.GetAlbumResponse, error) {
	result, err := clients.albums.GetAlbum(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	getAlbum(ctx *fasthttp.RequestCtx) {
	req := &pictures.GetAlbumRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := getAlbumGRPC(req)
	resolve(ctx, data.GetAlbum(), err, 401)
}
