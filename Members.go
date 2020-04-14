/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Thursday 09 January 2020 - 19:39:17
** @Filename:				Members.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Wednesday 01 April 2020 - 12:04:38
*******************************************************************************/

package			main

import			"context"
import			"encoding/json"
import			"github.com/microgolang/logs"
import			"github.com/valyala/fasthttp"
import			"github.com/panghostlin/SDK/Members"

type	sKeys struct {
	PrivateKey	string
	PublicKey	string
	PrivateIV	string
	PrivateSalt	string
}

func	setCookieAndResolve(ctx *fasthttp.RequestCtx, accessToken, publicKey, privateKey, privateIV, privateSalt string, err error) {
	if (err != nil) {
		ctx.Response.SetStatusCode(500)
		json.NewEncoder(ctx).Encode(false)
		return
	}
	keys := &sKeys{
		PrivateKey: privateKey,
		PublicKey: publicKey,
		PrivateIV: privateIV,
		PrivateSalt: privateSalt,
	}

	setCookie(ctx, `accessToken`, accessToken)
	ctx.Response.Header.SetContentType(`application/json`)
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(keys)
}
/******************************************************************************
**	Router proxy function to create a member.
**	Call the Members Microservice to create a new member.
******************************************************************************/
func	createNewMember(ctx *fasthttp.RequestCtx) {
	req := &members.CreateMemberRequest{}
	json.Unmarshal(ctx.PostBody(), &req)

	result, err := clients.members.CreateMember(context.Background(), req)
	if (err != nil) {
		logs.Error(`LoginMember : fail to login member`, err)
		setCookieAndResolve(ctx, ``, ``, ``, ``, ``, err)
	} else {
		setCookieAndResolve(ctx, result.GetAccessToken().Value, result.GetKeys().GetPublicKey(), result.GetKeys().GetPrivateKey(), result.GetKeys().GetPrivateIV(), result.GetKeys().GetPrivateSalt(), nil)
	}

}

/******************************************************************************
**	Router proxy function to login a member.
**	Call the Members Microservice to login an existing member.
******************************************************************************/
func	loginMember(ctx *fasthttp.RequestCtx) {
	req := &members.LoginMemberRequest{}
	json.Unmarshal(ctx.PostBody(), &req)

	result, err := clients.members.LoginMember(context.Background(), req)

	if (err != nil) {
		logs.Error(`LoginMember : fail to login member`, err)
		setCookieAndResolve(ctx, ``, ``, ``, ``, ``, err)
	} else {
		setCookieAndResolve(ctx, result.GetAccessToken().Value, result.GetKeys().GetPublicKey(), result.GetKeys().GetPrivateKey(), result.GetKeys().GetPrivateIV(), result.GetKeys().GetPrivateSalt(), nil)
	}
}

/******************************************************************************
**	CheckMember
**	Router proxy function to check if a member exists and is connected
******************************************************************************/
func	checkMember(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(200)
	json.NewEncoder(ctx).Encode(true)
}

/******************************************************************************
**	getMember
**	Router proxy function to retreive informations about the member
******************************************************************************/
func	getMember(ctx *fasthttp.RequestCtx) {
	memberID := ctx.UserValue("memberID").(string)
	req := &members.GetMemberRequest{MemberID: memberID}

	result, err := clients.members.GetMember(context.Background(), req)
	resolve(ctx, result, err, 401)
}
