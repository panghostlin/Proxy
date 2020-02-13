/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Sunday 05 January 2020 - 17:57:40
** @Filename:				main.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 13 February 2020 - 12:29:01
*******************************************************************************/

package			main

import			"os"
import			"log"
import			"crypto/tls"
import			"crypto/x509"
import			"io/ioutil"
import			"google.golang.org/grpc"
import			"google.golang.org/grpc/credentials"
import			"github.com/microgolang/logs"
import			"github.com/panghostlin/SDK/Keys"
import			"github.com/panghostlin/SDK/Members"
import			"github.com/panghostlin/SDK/Pictures"
import			"github.com/valyala/fasthttp"
import			"github.com/lab259/cors"

import		"net/http"


const	DEFAULT_CHUNK_SIZE = 64 * 1024
type	Sclients struct {
	members		members.MembersServiceClient
	keys		keys.KeysServiceClient
	pictures	pictures.PicturesServiceClient
	albums		pictures.AlbumsServiceClient
}
var		connections map[string](*grpc.ClientConn)
var		clients = &Sclients{}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func	serveProxy() {
	// crt := `/env/proxy.crt`
    // key := `/env/proxy.key`
	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		// AllowedOrigins: []string{
		// 	"https://api.majorcalamity.com", //PROD ADMIN
		// 	"https://majorcalamity.com", //PROD WEBSITE
		// },

		AllowedMethods: []string{`GET`, `POST`, `DELETE`, `PUT`, `OPTIONS`},
		AllowedHeaders:	[]string{
			`Origin`,
			`Access-Control-Allow-Origin`,
			`Access-Control-Allow-Credentials`,
			`Content-Type`,
			`Transfer-Encoding`,
			`Authorization`,
			`X-Content-Type`,
			`X-Content-Length`,
			`X-Content-Name`,
			`X-Content-Parts`,
			`X-Content-Last-Modified`,
			`X-Content-UUID`,
			`X-Content-AlbumID`,
			`X-Chunk-ID`,
		},
		ExposedHeaders: []string{`Set-Cookie`, `set-cookie`, `cookie`, `Content-Length`, `Content-Range`},
		AllowCredentials: true,
		OptionsPassthrough: true,
	})

	handler := c.Handler(InitRouter())
	// fasthttp.ListenAndServe(`:8000`, handler)
	_ = handler

	go func() {
		logs.Success(`Listening sockets on :8001`)
		log.Fatal(http.ListenAndServe(`:8001`, InitWebsocketRouter()))
	}()
	logs.Success(`Listening on :8000`)
	fasthttp.ListenAndServe(`:8000`, InitRouter())
}
func	bridgeInsecureMicroservice(serverName string, clientMS string) (*grpc.ClientConn) {
	logs.Warning("Using insecure connection")
	conn, err := grpc.Dial(serverName, grpc.WithInsecure())
    if err != nil {
		logs.Error("Did not connect", err)
		return nil
	}

	if (clientMS == `members`) {
		clients.members = members.NewMembersServiceClient(conn)
	} else if (clientMS == `keys`) {
		clients.keys = keys.NewKeysServiceClient(conn)
	} else if (clientMS == `pictures`) {
		clients.pictures = pictures.NewPicturesServiceClient(conn)
		clients.albums = pictures.NewAlbumsServiceClient(conn)
	}

	return conn
}
func	bridgeMicroservice(serverName string, clientMS string) (*grpc.ClientConn){
	crt := `/env/client.crt`
    key := `/env/client.key`
	caCert  := `/env/ca.crt`

    // Load the client certificates from disk
    certificate, err := tls.LoadX509KeyPair(crt, key)
    if err != nil {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
    }

    // Create a certificate pool from the certificate authority
    certPool := x509.NewCertPool()
    ca, err := ioutil.ReadFile(caCert)
    if err != nil {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
    }

    // Append the certificates from the CA
    if ok := certPool.AppendCertsFromPEM(ca); !ok {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
    }

    creds := credentials.NewTLS(&tls.Config{
        ServerName:   serverName,
        Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		InsecureSkipVerify: true,
    })

	conn, err := grpc.Dial(serverName, grpc.WithTransportCredentials(creds))
    if err != nil {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
	}

	if (clientMS == `members`) {
		clients.members = members.NewMembersServiceClient(conn)
	} else if (clientMS == `keys`) {
		clients.keys = keys.NewKeysServiceClient(conn)
	} else if (clientMS == `pictures`) {
		clients.pictures = pictures.NewPicturesServiceClient(conn)
		clients.albums = pictures.NewAlbumsServiceClient(conn)
	}

	return conn
}

func	main()	{
	connections = make(map[string](*grpc.ClientConn))
	connections[`members`] = bridgeMicroservice(`panghostlin-members:8010`, `members`)
	connections[`pictures`] = bridgeMicroservice(`panghostlin-pictures:8012`, `pictures`)

	serveProxy()
}
