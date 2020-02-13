/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:39:17
** @Filename:				Members.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 18:45:57
*******************************************************************************/

package			main

import			"context"
import			"encoding/json"
import			"github.com/microgolang/logs"
import			"github.com/valyala/fasthttp"
import			"github.com/panghostlin/SDK/Members"

/******************************************************************************
**	createMemberGRPC
**	Call the Members Microservice to create a new member.
**
**	CreateNewMember
**	Router proxy function to create a member.
******************************************************************************/
func	createMemberGRPC(data []byte) (string, *members.Cookie, string, bool, error) {
	req := &members.CreateMemberRequest{}
	json.Unmarshal(data, &req)

	result, err := clients.members.CreateMember(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return ``, &members.Cookie{}, ``, false, err
	}
	if (!result.Success) {
		logs.Error(`Failed to create this user`)
		return ``, &members.Cookie{}, ``, false, nil
	}
	return result.MemberID, result.AccessToken, result.GetHashKey(), true, nil
}
func	CreateNewMember(ctx *fasthttp.RequestCtx) {
	_, cookie, hashkey, success, err := createMemberGRPC(ctx.PostBody())
	if (!success || err != nil) {
		ctx.Response.SetStatusCode(500)
		json.NewEncoder(ctx).Encode(false)
		return
	}

	setCookie(ctx, `accessToken`, cookie.Value)
	setCookie(ctx, `hashKey`, hashkey)
	
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(success)
}


/******************************************************************************
**	loginMemberGRPC
**	Call the Members Microservice to login an existing member.
**
**	LoginMember
**	Router proxy function to login a member.
******************************************************************************/
func	loginMemberGRPC(data []byte) (string, *members.Cookie, string, bool, error) {
	req := &members.LoginMemberRequest{}
	json.Unmarshal(data, &req)

	result, err := clients.members.LoginMember(context.Background(), req)
	if (err != nil) {
		logs.Error(`Fail to communicate with microservice`, err)
		return ``, &members.Cookie{}, ``, false, err
	}
	if (!result.Success) {
		logs.Error(`Failed to login this user`)
		return ``, &members.Cookie{}, ``, false, nil
	}
	return result.GetMemberID(), result.GetAccessToken(), result.GetHashKey(), true, nil
}
func	LoginMember(ctx *fasthttp.RequestCtx) {
	_, cookie, hashkey, success, err := loginMemberGRPC(ctx.PostBody())
	if (!success || err != nil) {
		ctx.Response.SetStatusCode(500)
		json.NewEncoder(ctx).Encode(false)
		return
	}

	setCookie(ctx, `accessToken`, cookie.Value)
	setCookie(ctx, `hashKey`, hashkey)

	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(success)
}

/******************************************************************************
**	CheckMember
**	Router proxy function to check if a member exists and is connected
******************************************************************************/
func	CheckMember(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(true)
}