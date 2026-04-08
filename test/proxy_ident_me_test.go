package test

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	cronet "github.com/sagernet/cronet-go"

	"github.com/stretchr/testify/require"
)

// TestIdentMeViaHTTPProxyIPv6 checks that CronetProxy routes HTTPS through an HTTP proxy.
//
// Requires: network, libcronet.so on LD_LIBRARY_PATH, -tags with_purego.
//
//	export CRONET_PROXY_TEST_URL='http://user:pass@host:port'
//	export CRONET_PROXY_TEST_EXPECTED_IP='egress-ip-from-ident.me'
//	go test -tags with_purego -run TestIdentMeViaHTTPProxyIPv6 -v .
func TestIdentMeViaHTTPProxyIPv6(t *testing.T) {
	proxyURL := os.Getenv("CRONET_PROXY_TEST_URL")
	wantOutbound := strings.TrimSpace(os.Getenv("CRONET_PROXY_TEST_EXPECTED_IP"))
	if proxyURL == "" || wantOutbound == "" {
		t.Skip("set CRONET_PROXY_TEST_URL and CRONET_PROXY_TEST_EXPECTED_IP (see test comment)")
	}

	params := cronet.NewEngineParams()
	defer params.Destroy()

	require.NoError(t, params.SetCronetProxyURLs([]string{proxyURL}))
	params.SetUserAgent("cronet-go-proxy-test/1.0")
	params.SetEnableQuic(false)
	params.SetEnableHTTP2(true)
	params.SetEnableBrotli(false)
	params.SetHTTPCacheMode(cronet.HTTPCacheModeDisabled)

	engine := cronet.NewEngine()
	defer func() {
		engine.Shutdown()
		engine.Destroy()
	}()

	require.Equal(t, cronet.ResultSuccess, engine.StartWithParams(params))

	rt := &cronet.RoundTripper{Engine: engine}
	client := &http.Client{Transport: rt}

	resp, err := client.Get("https://ident.me")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "ident.me status")
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	got := strings.TrimSpace(string(body))
	require.Equal(t, wantOutbound, got, "ident.me should report the proxy egress IP")
}
