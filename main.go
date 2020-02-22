/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Sunday 05 January 2020 - 17:57:40
** @Filename:				main.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Friday 21 February 2020 - 17:46:42
*******************************************************************************/

package			main

import			"os"
import			"crypto/tls"
import			"crypto/x509"
import			"io/ioutil"
import			"google.golang.org/grpc"
import			"google.golang.org/grpc/credentials"
import			"github.com/microgolang/logs"
import			"github.com/panghostlin/SDK/Members"
import			"github.com/panghostlin/SDK/Pictures"
import			"github.com/valyala/fasthttp"
import			"github.com/lab259/cors"


const	DEFAULT_CHUNK_SIZE = 64 * 1024
type	sClients struct {
	members		members.MembersServiceClient
	pictures	pictures.PicturesServiceClient
	albums		pictures.AlbumsServiceClient
}
var		connections map[string](*grpc.ClientConn)
var		clients = &sClients{}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func	serveProxy() {
	logs.Success(`Listening on :8000`)

	if (os.Getenv("IS_LOCAL") == `true`) {
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
				`X-Content-Key`,
				`X-Content-IV`,
			},
			ExposedHeaders: []string{`Set-Cookie`, `set-cookie`, `cookie`},
			AllowCredentials: true,
		})
	
		fasthttp.ListenAndServe(`:8000`, c.Handler(initRouter()))
	} else {	
		fasthttp.ListenAndServe(`:8000`, initRouter())
	}
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
