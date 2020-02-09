/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:45:17
** @Filename:				Pictures.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 06 February 2020 - 20:54:16
*******************************************************************************/

package			main

import			"io"
import			"context"
import			"strconv"
import			"net/url"
import			"github.com/microgolang/logs"
import			"gitlab.com/betterpiwigo/sdk/Pictures"
import			"github.com/valyala/fasthttp"
import			"github.com/fasthttp/websocket"
import			"encoding/json"
import			"bytes"

type	WSResponse struct {
	UUID		string
	Step		int8
	Picture		*pictures.ListPictures_Content
	IsSuccess	bool
}

func	streamWebsocketMessage(contentUUID string, step int8, picture *pictures.ListPictures_Content, isSuccess bool) {
	if wsConn, _, ok := rm.LoadWs(contentUUID); ok {
		response := &WSResponse{}
		response.Step = step
		response.UUID = contentUUID
		response.Picture = picture
		response.IsSuccess = isSuccess

		wsConn.WriteJSON(response)
	}
}

/******************************************************************************
**	downloadPictureGRPC
**	Call the Picture Microservice to download an image.
**
**	DownloadPicture
**	Router proxy function to download an image.
******************************************************************************/
func	uploadPictureGRPC(memberID string, file []byte, contentUUID, contentName, contentType, albumID, lastModified string) error {
	/**************************************************************************
	**	0. Init the data to send to the Pictures microservice
	**************************************************************************/
	req := &pictures.UploadPictureRequest{
		MemberID: memberID,
		AlbumID: albumID,
		Content: &pictures.UploadPictureRequest_Content{
			Name: contentName,
			Type: contentType,
			OriginalTime: lastModified,
		},
	}

	/**************************************************************************
	**	1. Open the stream to send Req & the file, cut in chunk
	**************************************************************************/
	stream, err := clients.pictures.UploadPicture(context.Background())
	if (err != nil) {
		return err
	}

	/**************************************************************************
	**	2. Chunk the file according to DEFAULT_CHUNK_SIZE (64 * 1000) and send
	**	the full message to the Pictures microservice
	**************************************************************************/
	fileSize := len(file)
	for currentByte := 0; currentByte < fileSize; currentByte += DEFAULT_CHUNK_SIZE {
		if currentByte + DEFAULT_CHUNK_SIZE > fileSize {
			req.Chunk = file[currentByte:fileSize]
		} else {
			req.Chunk = file[currentByte : currentByte + DEFAULT_CHUNK_SIZE]
		}
		if err := stream.Send(req); err != nil {
			return err
		}
	}

	/**************************************************************************
	**	3. Close the stream when it's done
	**************************************************************************/
	stream.CloseSend()

	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if (recv.GetPicture() != nil) {
			streamWebsocketMessage(contentUUID, int8(recv.GetStep()), recv.GetPicture(), recv.GetSuccess())
			break
		}
		streamWebsocketMessage(contentUUID, int8(recv.GetStep()), recv.GetPicture(), true)
	}
	return nil
}
func	UploadPicture(ctx *fasthttp.RequestCtx) {
	contentType := string(ctx.Request.Header.Peek(`X-Content-Type`))
	contentNameEncoded := string(ctx.Request.Header.Peek(`X-Content-Name`))
	contentChunkIDStr := string(ctx.Request.Header.Peek(`X-Chunk-ID`))
	contentPartsStr := string(ctx.Request.Header.Peek(`X-Content-Parts`))
	contentUUID := string(ctx.Request.Header.Peek(`X-Content-UUID`))
	contentAlbumID := string(ctx.Request.Header.Peek(`X-Content-AlbumID`))
	contentLastModified := string(ctx.Request.Header.Peek(`X-Content-Last-Modified`))

	contentChunkID, _ := strconv.Atoi(contentChunkIDStr)
	contentParts, _ := strconv.Atoi(contentPartsStr)
	
	contentName, err := url.QueryUnescape(contentNameEncoded)
	if err != nil {
		contentName = ``
	}

	if block, ok := rm.LoadContent(contentUUID); ok {
		currentBlock := block
		currentBlock[contentChunkID] = append(currentBlock[contentChunkID], ctx.Request.Body()...)
		rm.SetContent(contentUUID, currentBlock)
	} else {
		block = make([]([]byte), contentParts + 1)
		currentBlock := block
		currentBlock[contentChunkID] = append(currentBlock[contentChunkID], ctx.Request.Body()...)
		rm.SetContent(contentUUID, currentBlock)
	}

	
	if _, ok := rm.LoadLen(contentUUID); ok {
		rm.IncLen(contentUUID)
	} else {
		rm.InitLen(contentUUID)
		rm.IncLen(contentUUID)
	}
	

	if len, ok := rm.LoadLen(contentUUID); ok {
		if (len >= uint(contentParts)) {
			if blobArr, ok := rm.LoadContent(contentUUID); ok {
				blob := bytes.Join(blobArr, nil)
	
				streamWebsocketMessage(contentUUID, 2, nil, true)

				uploadPictureGRPC(ctx.UserValue(`memberID`).(string), blob, contentUUID, contentName, contentType, contentAlbumID, contentLastModified)

				if wsConn, _, ok := rm.LoadWs(contentUUID); ok {
					wsConn.Close()
					rm.Delete(contentUUID)
				}

			}
		}
	}
}

/******************************************************************************
**	WSUploadPicture
**	Websocket opened with the UploadPicture call, sending information to the
**	client about the current status (Step) :
**	1 : Opening websocket and sending UUID for this upload
**	2 : All the image has been received and loaded, we can send it for the
**		encryption
**	3 : The encryption is done, we can now save the image
**	4 : The image is saved, sending it's ID to work with if on the client side
******************************************************************************/
func	WSUploadPicture(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		response := &WSResponse{}
		response.Step = 1
		response.UUID, _ = generateUUID(32)
		rm.InitWs(response.UUID, conn)
		rm.InitLen(response.UUID)
		conn.WriteJSON(response)

		for {
			if _, isOpen, ok := rm.LoadWs(response.UUID); ok {
				if (!isOpen) {
					return
				}
			}
		}

	})
	if (err != nil) {
		logs.Error(`Impossible to upgrade connexion`)
		return
	}
}

/******************************************************************************
**	downloadPictureGRPC
**	Call the Picture Microservice to download an image.
**
**	DownloadPicture
**	Router proxy function to download an image.
******************************************************************************/
func	downloadPictureGRPC(pictureID, pictureSize, hashKey string) (*pictures.DownloadPictureResponse, error) {
	/**************************************************************************
	**	0. Init the data to send to the Pictures microservice
	**************************************************************************/
	req := &pictures.DownloadPictureRequest{
		PictureID: pictureID,
		PictureSize: pictureSize,
		HashKey: hashKey,
	}

	/**************************************************************************
	**	1. Open the stream to receive the data
	**************************************************************************/
	stream, err := clients.pictures.DownloadPicture(context.Background(), req)
	if (err != nil) {
		return nil, err
	}

	/**************************************************************************
	**	2. Init the element to receive the response
	**************************************************************************/
	blob := make([]byte, 0)
	resp := &pictures.DownloadPictureResponse{}


	/**************************************************************************
	**	3. Loop to get all the data
	**************************************************************************/
	for {
		select {
			case <-stream.Context().Done():
				return nil, stream.Context().Err()
			default:
		}

		receiver, err := stream.Recv()
		if err == io.EOF {
			resp.Chunk = blob
			stream.Context().Done()
			return resp, nil
		}
		if err != nil {
			logs.Error("receive error : ", err)
			continue
		}
		blob = append(blob, receiver.GetChunk()...)
	}
}
func	DownloadPicture(ctx *fasthttp.RequestCtx) {
	hashKey := ctx.UserValue("hashKey").([]byte)
	pictureID := ctx.UserValue("pictureID").(string)
	pictureSize := ctx.UserValue("pictureSize").(string)

	response, err := downloadPictureGRPC(pictureID, pictureSize, string(hashKey))
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(false)	
		return
	}
	ctx.Response.Header.SetContentType(response.GetContentType())
	ctx.Response.SetStatusCode(200)
	ctx.Write(response.GetChunk())
}

/******************************************************************************
**	deletePicturesGRPC
******************************************************************************/
func	deletePicturesGRPC(memberID string, picturesID []string) (bool, error) {
	/**************************************************************************
	**	0. Init the data to send to the Pictures microservice
	**************************************************************************/
	req := &pictures.DeletePicturesRequest{MemberID: memberID, PicturesID: picturesID}

	data, err := clients.pictures.DeletePictures(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return false, err
	}
	return data.GetSuccess(), nil
}
func	DeletePictures(ctx *fasthttp.RequestCtx) {
	type	Srequest struct {PicturesID []string}
	request := &Srequest{}
	json.Unmarshal(ctx.PostBody(), &request)

	memberID := ctx.UserValue("memberID").(string)
	isSuccess, err := deletePicturesGRPC(memberID, request.PicturesID)

	if (err != nil || !isSuccess) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(isSuccess)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(isSuccess)
}


/******************************************************************************
**	listPicturesByMemberGRPC
******************************************************************************/
func	ListPicturesByMemberGRPC(memberID string) (*pictures.ListPicturesByMemberIDResponse, error) {
	/**************************************************************************
	**	0. Init the data to send to the Pictures microservice
	**************************************************************************/
	req := &pictures.ListPicturesByMemberIDRequest{MemberID: memberID}

	result, err := clients.pictures.ListPicturesByMemberID(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	ListPicturesByMember(ctx *fasthttp.RequestCtx) {
	memberID := ctx.UserValue("memberID").(string)
	data, err := ListPicturesByMemberGRPC(memberID)

	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(false)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data.GetPictures())
}


/******************************************************************************
**	listPicturesByAlbumGRPC
******************************************************************************/
func	ListPicturesByAlbumGRPC(memberID, albumID string) (*pictures.ListPicturesByAlbumIDResponse, error) {
	/**************************************************************************
	**	0. Init the data to send to the Pictures microservice
	**************************************************************************/
	req := &pictures.ListPicturesByAlbumIDRequest{}
	req.MemberID = memberID
	req.AlbumID = albumID

	result, err := clients.pictures.ListPicturesByAlbumID(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return nil, err
	}
	return result, nil
}
func	ListPicturesByAlbum(ctx *fasthttp.RequestCtx) {
	type	Srequest struct {
		AlbumID	string
	}
	request := &Srequest{}

	json.Unmarshal(ctx.PostBody(), &request)
	memberID := ctx.UserValue("memberID").(string)

	data, err := ListPicturesByAlbumGRPC(memberID, request.AlbumID)
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(false)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(data.GetPictures())
}


/******************************************************************************
**	listPicturesByAlbumGRPC
******************************************************************************/
func	setPictureAlbumGRPC(memberID, albumID string, groupIDs []string) (bool, error) {
	/**************************************************************************
	**	0. Init the data to send to the Pictures microservice
	**************************************************************************/
	req := &pictures.SetPicturesAlbumRequest{}
	req.MemberID = memberID
	req.AlbumID = albumID
	req.GroupIDs = groupIDs

	result, err := clients.pictures.SetPicturesAlbum(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return false, err
	}
	return result.GetSuccess(), nil
}
func	SetPicturesAlbum(ctx *fasthttp.RequestCtx) {
	type	Srequest struct {
		AlbumID string
		GroupIDs []string
	}
	request := &Srequest{}
	json.Unmarshal(ctx.PostBody(), &request)

	memberID := ctx.UserValue("memberID").(string)
	isSuccess, err := setPictureAlbumGRPC(memberID, request.AlbumID, request.GroupIDs)
	if (err != nil) {
		ctx.Response.Header.SetContentType(`application/json`)
		ctx.Response.SetStatusCode(404)
		json.NewEncoder(ctx).Encode(false)	
		return
	}
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(isSuccess)
}
