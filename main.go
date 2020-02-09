/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Sunday 05 January 2020 - 17:57:40
** @Filename:				main.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Tuesday 04 February 2020 - 20:51:34
*******************************************************************************/

package			main

import			_ "os"
import			"log"
import			"crypto/tls"
import			"crypto/x509"
import			"io/ioutil"
import			"google.golang.org/grpc"
import			"google.golang.org/grpc/credentials"
import			"github.com/microgolang/logs"
import			"gitlab.com/betterpiwigo/sdk/Keys"
import			"gitlab.com/betterpiwigo/sdk/Members"
import			"gitlab.com/betterpiwigo/sdk/Pictures"
import			"github.com/valyala/fasthttp"
import			"github.com/lab259/cors"


type	Sclients struct {
	members		members.MembersServiceClient
	keys		keys.KeysServiceClient
	pictures	pictures.PicturesServiceClient
	albums		pictures.AlbumsServiceClient
}
var		connections map[string](*grpc.ClientConn)
var		clients = &Sclients{}

const	DEFAULT_CHUNK_SIZE = 64 * 1024


func	StartRouter() {
	crt := `/env/proxy.crt`
    key := `/env/proxy.key`
	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowedMethods: []string{`GET`, `POST`, `DELETE`, `PUT`, `OPTIONS`, `OPTION`},
		AllowedHeaders:	[]string{
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
		ExposedHeaders: []string{`Set-Cookie`, `set-cookie`, `cookie`},
		AllowCredentials: true,
	})

	go func() {
		handler := c.Handler(InitRouter())
		logs.Success(`Listening on :80`)
		fasthttp.ListenAndServe(`:80`, handler)
	}()

	handler := c.Handler(InitRouter())
	logs.Success(`Listening on :443`)
	fasthttp.ListenAndServeTLS(`:443`, crt, key, handler)
}

func	InitGRPC(serverName string, clientMS string) (*grpc.ClientConn){
	crt := `/env/client.crt`
    key := `/env/client.key`
	caCert  := `/env/ca.crt`

    // Load the client certificates from disk
    certificate, err := tls.LoadX509KeyPair(crt, key)
    if err != nil {
        log.Fatalf("could not load client key pair: %s", err)
    }

    // Create a certificate pool from the certificate authority
    certPool := x509.NewCertPool()
    ca, err := ioutil.ReadFile(caCert)
    if err != nil {
        log.Fatalf("could not read ca certificate: %s", err)
    }

    // Append the certificates from the CA
    if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
    }

    creds := credentials.NewTLS(&tls.Config{
        ServerName:   serverName, // NOTE: this is required!
        Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		InsecureSkipVerify: true,
    })

    // Create a connection with the TLS credentials
	conn, err := grpc.Dial(serverName, grpc.WithTransportCredentials(creds))
    if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}

	if (clientMS == `members`) {
		clients.members = members.NewMembersServiceClient(conn)
	} else if (clientMS == `keys`) {
		// clients.keys = keys.NewKeysServiceClient(conn)
	} else if (clientMS == `pictures`) {
		clients.pictures = pictures.NewPicturesServiceClient(conn)
		clients.albums = pictures.NewAlbumsServiceClient(conn)
	}

	return conn
}

func	main()	{
	logs.Pretty(generateNonce(32))


	connections = make(map[string](*grpc.ClientConn))
	connections[`members`] = InitGRPC(`piwigo-members:8010`, `members`)
	connections[`pictures`] = InitGRPC(`piwigo-pictures:8012`, `pictures`)
	// connections[`members`] = InitGRPC(`localhost:8010`, `members`)
	// connections[`pictures`] = InitGRPC(`localhost:8012`, `pictures`)

	StartRouter()
}
