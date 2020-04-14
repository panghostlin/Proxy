/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:45:17
** @Filename:				Pictures.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Wednesday 01 April 2020 - 12:04:34
*******************************************************************************/

package			main

import			"io"
import			"sync"
import			"context"
import			"strconv"
import			"io/ioutil"
import			"github.com/microgolang/logs"
import			"github.com/panghostlin/SDK/Pictures"
import			"github.com/valyala/fasthttp"
import			"github.com/fasthttp/websocket"
import			"encoding/json"
import			"bytes"

/******************************************************************************
**	wsResponse
**	Websocket to stream the picture status to the client.
**	- step: 1 => default
**	- step: 2 => sending image to the picture MS
**	- step: 3 => done saving the image
******************************************************************************/
type	wsResponse struct {
	UUID		string
	Step		int8
	Picture		*pictures.ListPictures_Content
	IsSuccess	bool
}

var		streamWebsockerMutex = sync.Mutex{}
func	streamWebsocketMessage(UUIDWithSize string, step int8, picture *pictures.ListPictures_Content, isSuccess bool) {
	if wsConn, _, ok := rm.loadWs(UUIDWithSize); ok {

		streamWebsockerMutex.Lock()
		defer streamWebsockerMutex.Unlock()
		
		response := &wsResponse{}
		response.Step = step
		response.UUID = UUIDWithSize
		response.Picture = picture
		response.IsSuccess = isSuccess

		err := wsConn.WriteJSON(response)
		if (err != nil) {
			logs.Warning(err.Error())
		}
	}
}

/******************************************************************************
**	downloadPictureGRPC
**	Call the Picture Microservice to download an image.
**
**	DownloadPicture
**	Router proxy function to download an image.
******************************************************************************/
func	uploadPictureGRPC(req *pictures.UploadPictureRequest, file []byte, contentUUID, isLast string) error {
	step := int8(3)
	if (isLast == `true`) {
		step = int8(4)
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

	/**************************************************************************
	**	We are creating a first ref in the mapping, with the default UUID, in
	**	order to know the UUID upload status for the entire batch
	**************************************************************************/
	for {
		recv, err := stream.Recv()
		if (err == io.EOF) {
			break
		} else if (err != nil) {
			streamWebsocketMessage(contentUUID, step, nil, false)
			break
		}

		if (recv.GetPicture() != nil) {
			streamWebsocketMessage(contentUUID, step, recv.GetPicture(), recv.GetSuccess())
			break
		} else {
			streamWebsocketMessage(contentUUID, step, recv.GetPicture(), recv.GetSuccess())
		}
	}
	if (step == 4) {
		if _, ok := rm.loadContent(contentUUID); !ok {
			rm.delete(contentUUID)
		}
	}
	return err
}

func	uploadPicture(ctx *fasthttp.RequestCtx) {
	/**************************************************************************
	**	Getting all the data from the client request (the file and the helpers)
	**************************************************************************/
	isLast := string(ctx.FormValue(`isLast`))
	contentUUID := string(ctx.FormValue(`fileUUID`))
	contentSizeType := string(ctx.FormValue(`fileSizeType`))
	contentChunkID, _ := strconv.Atoi(string(ctx.FormValue(`fileChunkID`)))
	contentParts, _ := strconv.Atoi(string(ctx.FormValue(`fileParts`)))
	file, _ := ctx.FormFile(`file`)
	fileContent, _ := file.Open()
	byteContainer, _ := ioutil.ReadAll(fileContent)

	/**************************************************************************
	**	We are creating this all upload reference, wich tell us if the UUID has
	**	been open, not the particular UUID_SIZE
	**************************************************************************/
	if isOpen, ok := rm.loadRefOpen(contentUUID); !ok || isOpen == false {
		streamWebsocketMessage(contentUUID, 2, nil, true)
		rm.setRefOpen(contentUUID, true)
	}

	/**************************************************************************
	**	We are creating a second ref in the mapping, for this specific image
	**	upload by contactaining the UUID with the contentSizeType.
	**	This will allow us to work with the upload status of this, and only
	**	this uploaded size.
	**************************************************************************/
	UUIDWithSize := contentUUID + `_` + contentSizeType
	if block, ok := rm.loadContent(UUIDWithSize); ok {
		logs.Pretty(`HERE`)
		currentBlock := block
		currentBlock[contentChunkID] = append(currentBlock[contentChunkID], byteContainer...)
		rm.setContent(UUIDWithSize, currentBlock)
	} else {
		rm.setWsOpen(UUIDWithSize, true)
		block = make([]([]byte), contentParts + 1)
		currentBlock := block
		currentBlock[contentChunkID] = append(currentBlock[contentChunkID], byteContainer...)
		rm.setContent(UUIDWithSize, currentBlock)
	}

	/**************************************************************************
	**	We are getting the current number of batch we got from the client.
	**	When this len === the content part, we got all the data to continue.
	**************************************************************************/
	if _, ok := rm.loadLen(UUIDWithSize); ok {
		rm.incLen(UUIDWithSize)
	} else {
		rm.initLen(UUIDWithSize)
		rm.incLen(UUIDWithSize)
	}


	if len, ok := rm.loadLen(UUIDWithSize); ok {
		if (len >= uint(contentParts)) {
			if blobArr, ok := rm.loadContent(UUIDWithSize); ok {
				blob := bytes.Join(blobArr, nil)
	

				fileWidthStr := string(ctx.FormValue(`fileWidth`))
				fileWidth, _ := strconv.Atoi(fileWidthStr)
				fileHeightStr := string(ctx.FormValue(`fileHeight`))
				fileHeight, _ := strconv.Atoi(fileHeightStr)
			
				req := &pictures.UploadPictureRequest{
					MemberID: ctx.UserValue(`memberID`).(string),
					AlbumID: string(ctx.FormValue(`fileAlbumID`)),
					Content: &pictures.UploadPictureRequest_Content{
						Name: string(ctx.FormValue(`fileName`)),
						Type: string(ctx.FormValue(`fileType`)),
						SizeType: contentSizeType,
						OriginalTime: string(ctx.FormValue(`fileLastModified`)),
						Width: int32(fileWidth), 
						Height: int32(fileHeight),
						GroupID: contentUUID,
					},
					Crypto: &pictures.PictureCrypto{
						Key: string(ctx.FormValue(`encryptionKey`)),
						IV: string(ctx.FormValue(`encryptionIV`)),
					},
				}

				uploadPictureGRPC(req, blob, contentUUID, isLast)

				rm.delete(UUIDWithSize)

				if (isLast == `true`) {
					if wsConn, _, ok := rm.loadWs(contentUUID); ok {
						wsConn.Close()
						rm.delete(contentUUID)
					}
				}
			}
		}
	}
}


/******************************************************************************
**	wsUploadPicture
**	Websocket opened with the UploadPicture call, sending information to the
**	client about the current status (Step) :
**	1 : Opening websocket and sending UUID for this upload
**	2 : All the image has been received and loaded, we can send it for the
**		encryption
**	3 : The encryption is done, we can now save the image
**	4 : The image is saved, sending it's ID to work with if on the client side
******************************************************************************/
func	wsUploadPicture(ctx *fasthttp.RequestCtx) {
	fileUUID := ctx.UserValue("fileUUID")
	err := fastupgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		response := &wsResponse{}
		response.Step = 1
		if (fileUUID == nil) {
			response.UUID, _ = generateUUID(32)
		} else {
			response.UUID = fileUUID.(string)
		}
		rm.initWs(response.UUID, conn)
		rm.initLen(response.UUID)
		err := conn.WriteJSON(response)
		if (err != nil) {
			logs.Warning(err.Error())
		}

		for {
			if _, isOpen, ok := rm.loadWs(response.UUID); ok {
				if (!isOpen) {
					return
				}
			}
		}

	})
	if (err != nil) {
		logs.Error(`Impossible to upgrade connexion : ` + err.Error())
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
			stream.Context().Done()
			return resp, nil
		}
		if err != nil {
			logs.Error("receive error : ", err)
			continue
		}
		resp.Crypto = receiver.GetCrypto()
		resp.Crypto.Key = receiver.GetCrypto().GetKey()
		resp.Crypto.IV = receiver.GetCrypto().GetIV()
		resp.ContentType = receiver.GetContentType()
		resp.Chunk = append(resp.GetChunk(), receiver.GetChunk()...)
	}
}
func	downloadPicture(ctx *fasthttp.RequestCtx) {
	hashKey := ctx.UserValue("hashKey").([]byte)
	pictureID := ctx.UserValue("pictureID").(string)
	pictureSize := ctx.UserValue("pictureSize").(string)

	response, err := downloadPictureGRPC(pictureID, pictureSize, string(hashKey))
	resolvePicture(ctx, response, err)
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
func	deletePictures(ctx *fasthttp.RequestCtx) {
	type	Srequest struct {PicturesID []string}
	request := &Srequest{}
	json.Unmarshal(ctx.PostBody(), &request)

	memberID := ctx.UserValue("memberID").(string)
	isSuccess, err := deletePicturesGRPC(memberID, request.PicturesID)
	resolve(ctx, isSuccess, err, 401)
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
func	listPicturesByMember(ctx *fasthttp.RequestCtx) {
	memberID := ctx.UserValue("memberID").(string)
	data, err := ListPicturesByMemberGRPC(memberID)
	resolve(ctx, data.GetPictures(), err, 401)
}


/******************************************************************************
**	listPicturesByAlbumGRPC
******************************************************************************/
func	listPicturesByAlbumGRPC(memberID, albumID string) (*pictures.ListPicturesByAlbumIDResponse, error) {
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
func	listPicturesByAlbum(ctx *fasthttp.RequestCtx) {
	type	Srequest struct {AlbumID string}
	request := &Srequest{}

	json.Unmarshal(ctx.PostBody(), &request)
	memberID := ctx.UserValue("memberID").(string)

	data, err := listPicturesByAlbumGRPC(memberID, request.AlbumID)
	resolve(ctx, data.GetPictures(), err, 401)
}


/******************************************************************************
**	setPictureAlbumGRPC
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
func	setPicturesAlbum(ctx *fasthttp.RequestCtx) {
	type	Srequest struct {
		AlbumID string
		GroupIDs []string
	}
	request := &Srequest{}
	json.Unmarshal(ctx.PostBody(), &request)

	memberID := ctx.UserValue("memberID").(string)
	isSuccess, err := setPictureAlbumGRPC(memberID, request.AlbumID, request.GroupIDs)
	resolve(ctx, isSuccess, err, 401)
}

/******************************************************************************
**	setPicturesDateGRPC
******************************************************************************/
func	setPicturesDateGRPC(memberID, newDate string, groupIDs []string) (bool, error) {
	/**************************************************************************
	**	0. Init the data to send to the Pictures microservice
	**************************************************************************/
	req := &pictures.SetPicturesDateRequest{
		MemberID: memberID,
		NewDate: newDate,
		GroupIDs: groupIDs,
	}
	/**************************************************************************
	**	1. Send the data to the microservice
	**************************************************************************/
	result, err := clients.pictures.SetPicturesDate(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return false, err
	}
	return result.GetSuccess(), nil
}
func	setPicturesDate(ctx *fasthttp.RequestCtx) {
	type	Srequest struct {
		NewDate string
		GroupIDs []string
	}
	request := &Srequest{}
	json.Unmarshal(ctx.PostBody(), &request)

	memberID := ctx.UserValue("memberID").(string)
	isSuccess, err := setPicturesDateGRPC(memberID, request.NewDate, request.GroupIDs)
	resolve(ctx, isSuccess, err, 401)
}
