package common

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/buildkit/client"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// ResolveClient resolves a client from CLI args
func ResolveClient(c *cli.Context) (*client.Client, error) {
	serverName := c.GlobalString("tlsservername")
	if serverName == "" {
		// guess servername as hostname of target address
		uri, err := url.Parse(c.GlobalString("addr"))
		if err != nil {
			return nil, err
		}
		serverName = uri.Hostname()
	}

	var caCert string
	var cert string
	var key string

	tlsDir := c.GlobalString("tlsdir")

	if tlsDir != "" {
		// Look for ca.pem and, if it exists, set caCert to that
		// Look for cert.pem and, if it exists, set key to that
		// Look for key.pem and, if it exists, set tlsDir to that
		for _, v := range [3]string{"ca.pem", "cert.pem", "key.pem"} {
			file := filepath.Join(tlsDir, v)
			if _, err := os.Stat(file); err == nil {
				switch v {
				case "ca.pem":
					caCert = file
				case "cert.pem":
					cert = file
				case "key.pem":
					key = file
				}
			} else {
				return nil, err
			}
		}

		if c.GlobalString("tlscacert") != "" || c.GlobalString("tlscert") != "" || c.GlobalString("tlskey") != "" {
			return nil, errors.New("cannot specify tlsdir and tlscacert/tlscert/tlskey at the same time")
		}
	} else {
		caCert = c.GlobalString("tlscacert")
		cert = c.GlobalString("tlscert")
		key = c.GlobalString("tlskey")
	}

	opts := []client.ClientOpt{client.WithFailFast()}

	ctx := CommandContext(c)

	if span := opentracing.SpanFromContext(ctx); span != nil {
		opts = append(opts, client.WithTracer(span.Tracer()))
	}

	if caCert != "" || cert != "" || key != "" {
		opts = append(opts, client.WithCredentials(serverName, caCert, cert, key))
	}

	timeout := time.Duration(c.GlobalInt("timeout"))
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	return client.New(ctx, c.GlobalString("addr"), opts...)
}
