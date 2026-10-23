package service

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuditImageDownloadRejectsMulticastAndBroadcastBeforeTransport(t *testing.T) {
	for _, target := range []string{"http://224.0.0.1/image.png", "http://239.255.255.250/image.png", "http://255.255.255.255/image.png", "http://[ff0e::1]/image.png"} {
		t.Run(target, func(t *testing.T) {
			for _, proxyURL := range []*url.URL{nil, {Scheme: "http", Host: "127.0.0.1:7890"}} {
				transport, directCalls, proxiedCalls := newCountingImageDownloadTransport(proxyURL)
				req, err := http.NewRequest(http.MethodGet, target, nil)
				require.NoError(t, err)
				_, err = transport.RoundTrip(req)
				require.Error(t, err)
				require.Zero(t, directCalls.Load())
				require.Zero(t, proxiedCalls.Load())
				blocked, err := isNonPublicImageDownloadHost(context.Background(), req.URL.Hostname())
				require.NoError(t, err)
				require.True(t, blocked, "the proxy's host check must also enforce unicast")
			}
		})
	}
	for _, address := range []string{"224.0.0.1:443", "239.255.255.250:443", "255.255.255.255:443", "[ff0e::1]:443"} {
		require.Error(t, imageDownloadDialControl("tcp", address, nil), "the actual socket destination must also be unicast")
	}
}
