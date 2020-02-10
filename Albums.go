/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 14 January 2020 - 20:21:56
** @Filename:				Albums.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Monday 10 February 2020 - 11:56:39
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
			CoverPicture0ID: req.GetCoverPicture0ID(),
			CoverPicture1ID: req.GetCoverPicture1ID(),
			CoverPicture2ID: req.GetCoverPicture2ID(),
			Pictures: req.GetPictures(),
		},
	)

	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	CreateAlbum(ctx *fasthttp.RequestCtx) {
	req := &pictures.CreateAlbumRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	memberID := ctx.UserValue("memberID").(string)

	data, err := createAlbumGRPC(memberID, req)

	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(false)
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data)
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
func	ListAlbums(ctx *fasthttp.RequestCtx) {
	memberID := ctx.UserValue("memberID").(string)
	data, err := listAlbumsGRPC(memberID)

	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(true)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data.GetAlbums())
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
func	SetAlbumCover(ctx *fasthttp.RequestCtx) {
	req := &pictures.SetAlbumCoverRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := setAlbumCoverGRPC(req)
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(true)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data.GetAlbumID())
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
func	SetAlbumName(ctx *fasthttp.RequestCtx) {
	req := &pictures.SetAlbumNameRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := setAlbumNameGRPC(req)
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(true)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data.GetAlbumID())
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
func	DeleteAlbum(ctx *fasthttp.RequestCtx) {
	req := &pictures.DeleteAlbumRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := deleteAlbumGRPC(req)
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(true)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data.GetSuccess())
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
func	GetAlbum(ctx *fasthttp.RequestCtx) {
	req := &pictures.GetAlbumRequest{}
	json.Unmarshal(ctx.PostBody(), &req)
	req.MemberID = ctx.UserValue("memberID").(string)

	data, err := getAlbumGRPC(req)
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(true)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data.GetAlbum())
}
