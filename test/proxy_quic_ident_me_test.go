package test

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	cronet "github.com/dkrhodes/cronet-go"

	"github.com/stretchr/testify/require"
)

// TestIdentMeViaQUICProxy checks that a Cronet QUIC (MASQUE) proxy is used for HTTPS
// to ident.me with QUIC enabled to the origin when the server offers it (Alt-Svc).
// Egress IP must match the proxy exit.
//
// Requires: network, libcronet with quic:// CronetProxy support, -tags with_purego.
//
//	export CRONET_QUIC_PROXY_TEST_URL='quic://user-1000:pass@127.0.0.1:4142'
//	export CRONET_QUIC_PROXY_TEST_EXPECTED_IP='<egress-ip-seen-by-ident.me>'
//
// Trust for the proxy TLS certificate (pick one):
//
//	export CRONET_QUIC_PROXY_TEST_CA_PEM=/path/to/ca.pem
//
// dev-only alternative:
//
//	export CRONET_QUIC_PROXY_TEST_INSECURE=1
//
//	go test -tags with_purego -run TestIdentMeViaQUICProxy -v .
func TestIdentMeViaQUICProxy(t *testing.T) {
	proxyURL := strings.TrimSpace(os.Getenv("CRONET_QUIC_PROXY_TEST_URL"))
	wantOutbound := strings.TrimSpace(os.Getenv("CRONET_QUIC_PROXY_TEST_EXPECTED_IP"))
	caPath := strings.TrimSpace(os.Getenv("CRONET_QUIC_PROXY_TEST_CA_PEM"))
	insecure := strings.TrimSpace(os.Getenv("CRONET_QUIC_PROXY_TEST_INSECURE")) == "1"

	if proxyURL == "" || wantOutbound == "" {
		t.Skip("set CRONET_QUIC_PROXY_TEST_URL and CRONET_QUIC_PROXY_TEST_EXPECTED_IP (see test comment)")
	}
	if !insecure && caPath == "" {
		t.Skip("set CRONET_QUIC_PROXY_TEST_CA_PEM or CRONET_QUIC_PROXY_TEST_INSECURE=1 for proxy TLS trust")
	}
	if insecure && caPath != "" {
		t.Fatal("use only one of CRONET_QUIC_PROXY_TEST_CA_PEM or CRONET_QUIC_PROXY_TEST_INSECURE")
	}

	var caPEM string
	if caPath != "" {
		b, err := os.ReadFile(caPath)
		require.NoError(t, err, "read CA PEM")
		caPEM = string(b)
	}

	params := cronet.NewEngineParams()
	defer params.Destroy()

	require.NoError(t, params.SetCronetProxyURLs([]string{proxyURL}))
	params.SetUserAgent("cronet-go-quic-proxy-ident/1.0")
	params.SetEnableQuic(true)
	params.SetEnableHTTP2(true)
	params.SetEnableBrotli(false)
	params.SetHTTPCacheMode(cronet.HTTPCacheModeDisabled)
	params.SetEnablePublicKeyPinningBypassForLocalTrustAnchors(true)

	engine := cronet.NewEngine()
	defer func() {
		engine.Shutdown()
		engine.Destroy()
	}()

	if insecure {
		require.True(t, engine.SetInsecureSkipVerify(), "SetInsecureSkipVerify")
	} else {
		require.True(t, engine.SetTrustedRootCertificates(caPEM), "SetTrustedRootCertificates")
	}

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
