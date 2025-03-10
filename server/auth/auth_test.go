package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"testing"

	"github.com/pinterest/knox"
)

func TestUserCanAccess(t *testing.T) {
	u := NewUser("test", []string{"returntrue"})
	a1 := knox.Access{ID: "test", AccessType: knox.Write, Type: knox.User}
	a2 := knox.Access{ID: "returnfalse", AccessType: knox.Admin, Type: knox.UserGroup}
	a3 := knox.Access{ID: "returntrue", AccessType: knox.Admin, Type: knox.UserGroup}

	acl1 := knox.ACL([]knox.Access{a1})
	if !u.CanAccess(acl1, knox.Write) {
		t.Error("user can't access user permission matching id")
	}
	if u.CanAccess(acl1, knox.Admin) {
		t.Error("user can access user permission with increased access type")
	}

	acl2 := knox.ACL([]knox.Access{a2})
	if u.CanAccess(acl2, knox.Write) {
		t.Error("user can access group they are not in")
	}

	acl3 := knox.ACL([]knox.Access{a3})
	if !u.CanAccess(acl3, knox.Read) {
		t.Error("user can't access group they are in")
	}

	acl4 := knox.ACL([]knox.Access{a2, a3, a1})
	if !u.CanAccess(acl4, knox.Admin) {
		t.Error("user can't access group they are in with multiple members")
	}

	acl5 := knox.ACL([]knox.Access{})
	if u.CanAccess(acl5, knox.Admin) {
		t.Error("user can access empty ACL")
	}
}
func TestMachineCanAccess(t *testing.T) {
	u := machine("test001")
	a1 := knox.Access{ID: "test001", AccessType: knox.Write, Type: knox.Machine}
	a2 := knox.Access{ID: "doesnotmatch", AccessType: knox.Admin, Type: knox.MachinePrefix}
	a3 := knox.Access{ID: "test", AccessType: knox.Admin, Type: knox.MachinePrefix}

	acl1 := knox.ACL([]knox.Access{a1})
	if !u.CanAccess(acl1, knox.Write) {
		t.Error("machine can't access user permission matching id")
	}
	if u.CanAccess(acl1, knox.Admin) {
		t.Error("machine can access user permission with increased access type")
	}

	acl2 := knox.ACL([]knox.Access{a2})
	if u.CanAccess(acl2, knox.Write) {
		t.Error("machine can access group they are not in")
	}

	acl3 := knox.ACL([]knox.Access{a3})
	if !u.CanAccess(acl3, knox.Read) {
		t.Error("machine can't access group they are in")
	}

	acl4 := knox.ACL([]knox.Access{a2, a3, a1})
	if !u.CanAccess(acl4, knox.Admin) {
		t.Error("machine can't access group they are in with multiple members")
	}

	acl5 := knox.ACL([]knox.Access{})
	if u.CanAccess(acl5, knox.Admin) {
		t.Error("machine can access empty ACL")
	}
}

func TestServiceCanAccess(t *testing.T) {
	s := NewService("example.com", "serviceA")
	a1 := knox.Access{ID: "spiffe://example.com/serviceA", AccessType: knox.Read,
		Type: knox.Service}

	acl1 := knox.ACL([]knox.Access{a1})
	if s.CanAccess(acl1, knox.Admin) {
		t.Error("service can't access user permission matching id")
	}
	if !s.CanAccess(acl1, knox.Read) {
		t.Error("service can't access group they are in")
	}
}

func TestPrincipalMuxType(t *testing.T) {
	u := NewUser("test", []string{"returntrue"})
	s := NewService("example.com", "serviceA")

	user := knox.NewPrincipalMux(u, map[string]knox.Principal{"foo": u})
	service := knox.NewPrincipalMux(s, map[string]knox.Principal{"foo": s})
	both := knox.NewPrincipalMux(u, map[string]knox.Principal{"foo": u, "bar": s})

	if user.Type() != "user" {
		t.Error("Type of user-only mux should be 'user'")
	}
	if service.Type() != "service" {
		t.Error("Type of service-only mux should be 'service'")
	}
	if !strings.Contains(both.Type(), "mux") || !strings.Contains(both.Type(), "user") || !strings.Contains(both.Type(), "service") {
		t.Error("Type of mux with both should contain both user and service in the type string")
	}
}

func TestPrincipalMuxUserOrService(t *testing.T) {
	u := NewUser("test", []string{"returntrue"})
	s := NewService("example.com", "serviceA")
	userMux := knox.NewPrincipalMux(u, map[string]knox.Principal{"foo": u, "bar": s})
	serviceMux := knox.NewPrincipalMux(s, map[string]knox.Principal{"foo": u, "bar": s})

	if !IsUser(userMux) {
		t.Error("IsUser failed to identify that mux is user first.")
	}
	if IsService(userMux) {
		t.Error("IsService failed to identify that mux is user first, and thus not a service.")
	}
	if IsUser(serviceMux) {
		t.Error("IsUser failed to identify that mux is service first, and thus not a user.")
	}
	if !IsService(serviceMux) {
		t.Error("IsService failed to identify that mux is service first.")
	}
}

const caCert = `-----BEGIN CERTIFICATE-----
MIICOjCCAeCgAwIBAgIUIKkBZQbtx8rVaWIOhpabkqZSqecwCgYIKoZIzj0EAwIw
aTELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xHzAdBgNVBAoTFkludGVybmV0IFdpZGdldHMsIEluYy4xDDAK
BgNVBAsTA1dXVzAeFw0xNjA0MTMwMDQ2MDBaFw0yMTA0MTIwMDQ2MDBaMGkxCzAJ
BgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJh
bmNpc2NvMR8wHQYDVQQKExZJbnRlcm5ldCBXaWRnZXRzLCBJbmMuMQwwCgYDVQQL
EwNXV1cwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAT2bdxWXCY2y7YwFdDqgBOG
xMk66tC6L2CXBlSue9/GNypv771VfRFf7RtNa/3/pTLEVJZaj25sOXCJaNeFX9GE
o2YwZDAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/BAgwBgEB/wIBAjAdBgNVHQ4E
FgQUHeijCT+fV2y15OFqaV+BuROvrG4wHwYDVR0jBBgwFoAUHeijCT+fV2y15OFq
aV+BuROvrG4wCgYIKoZIzj0EAwIDSAAwRQIgZpgo1bmCAdSaVCqJKDmMKfui2dT/
3ucYcCZi9dUZjtMCIQC/d1se0XhhZ8eRfqzf0Uj0jHvan4opB0aD5CgSVlct0w==
-----END CERTIFICATE-----`

// clientCertB64 is a base64 encoded cert (inner contents of a CERTIFICATE pem block)
const clientCertB64 = `MIICjzCCAjSgAwIBAgIUUOdxnpGiNZhsB0AySQMJ+Lx5WqEwCgYIKoZIzj0EAwIw
aTELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xHzAdBgNVBAoTFkludGVybmV0IFdpZGdldHMsIEluYy4xDDAK
BgNVBAsTA1dXVzAeFw0xNjA0MTMwMDQ5MDBaFw0xNzA0MTMwMDQ5MDBaMIGDMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzEWMBQGA1UEChMNUGludGVyZXN0IEluYzETMBEGA1UECxMKcGluMjIw
LmNvbTEaMBgGA1UEAxMRZGV2LWRldmlubHVuZGJlcmcwWTATBgcqhkjOPQIBBggq
hkjOPQMBBwNCAATntqtFfYd8qDNzhRgbi3CMgRHnz92Vt70nmJntO6XIk1tHyP9b
k4aJe7KplZG8Fc56lB4Y+kE+6INE/6OkYqxKo4GeMIGbMA4GA1UdDwEB/wQEAwIF
oDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAd
BgNVHQ4EFgQUKRAflMKxR43xkv9w+yN7raOGFpAwHwYDVR0jBBgwFoAUHeijCT+f
V2y15OFqaV+BuROvrG4wHAYDVR0RBBUwE4IRZGV2LWRldmlubHVuZGJlcmcwCgYI
KoZIzj0EAwIDSQAwRgIhAIRd87po8pVE1CFSSM/uPwHFPg3gWlC1Pvl5h5e+2ogf
AiEA/GIpOpaFQbGSs42rKugOBngKtF0fuRAo2r4vMyL559A=`

func TestMTLSSuccess(t *testing.T) {
	hostname := "dev-devinlundberg"
	expected := "dev-devinlundberg"
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Authorization", "0t"+hostname)
	req.RemoteAddr = "0.0.0.0:23423"
	certBytes := make([]byte, base64.StdEncoding.DecodedLen(len(clientCertB64)))
	n, err := base64.StdEncoding.Decode(certBytes, []byte(clientCertB64))
	if err != nil {
		t.Fatal(err.Error())
	}
	c, _ := x509.ParseCertificate(certBytes[:n])
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{c},
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM([]byte(caCert))
	a := MTLSAuthProvider{
		CAs:  caPool,
		time: func() time.Time { return time.Date(2016, time.April, 22, 11, 0, 0, 0, time.UTC) },
	}
	p, err := a.Authenticate(hostname, req)

	if err != nil {
		t.Fatal(err.Error())
	}
	if p != machine(expected) {
		t.Fatal("Hostnames don't match got: " + p.GetID())
	}
}

func TestMTLSBadTime(t *testing.T) {
	hostname := "dev-devinlundberg"
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Authorization", "0t"+hostname)
	req.RemoteAddr = "0.0.0.0:23423"
	certBytes := make([]byte, base64.StdEncoding.DecodedLen(len(clientCertB64)))
	n, err := base64.StdEncoding.Decode(certBytes, []byte(clientCertB64))
	if err != nil {
		t.Fatal(err.Error())
	}
	c, _ := x509.ParseCertificate(certBytes[:n])
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{c},
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM([]byte(caCert))
	a := MTLSAuthProvider{
		CAs:  caPool,
		time: func() time.Time { return time.Date(2018, time.April, 22, 11, 0, 0, 0, time.UTC) },
	}
	_, err = a.Authenticate(hostname, req)

	if err == nil {
		t.Fatal("Call should fail due to expired cert")
	}
}

func TestMTLSNoCA(t *testing.T) {
	hostname := "dev-devinlundberg"
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Authorization", "0t"+hostname)
	req.RemoteAddr = "0.0.0.0:23423"
	certBytes := make([]byte, base64.StdEncoding.DecodedLen(len(clientCertB64)))
	n, err := base64.StdEncoding.Decode(certBytes, []byte(clientCertB64))
	if err != nil {
		t.Fatal(err.Error())
	}
	c, _ := x509.ParseCertificate(certBytes[:n])
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{c},
	}

	a := MTLSAuthProvider{
		CAs:  x509.NewCertPool(),
		time: func() time.Time { return time.Date(2016, time.April, 22, 11, 0, 0, 0, time.UTC) },
	}
	_, err = a.Authenticate(hostname, req)

	if err == nil {
		t.Fatal("There is no matching CA for this cert")
	}
}

func TestMTLSBadHostname(t *testing.T) {
	hostname := "BadHostname"
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Authorization", "0t"+hostname)
	req.RemoteAddr = "0.0.0.0:23423"
	certBytes := make([]byte, base64.StdEncoding.DecodedLen(len(clientCertB64)))
	n, err := base64.StdEncoding.Decode(certBytes, []byte(clientCertB64))
	if err != nil {
		t.Fatal(err.Error())
	}
	c, _ := x509.ParseCertificate(certBytes[:n])
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{c},
	}

	a := MTLSAuthProvider{
		CAs:  x509.NewCertPool(),
		time: func() time.Time { return time.Date(2016, time.April, 22, 11, 0, 0, 0, time.UTC) },
	}
	_, err = a.Authenticate(hostname, req)

	if err == nil {
		t.Fatal("hostname should not match")
	}

}

func TestGetUser(t *testing.T) {
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwicHJlZmVycmVkX3VzZXJuYW1lIjoidGVzdHVzZXIiLCJncm91cCI6InRlc3Rncm91cCIsImFkbWluIjp0cnVlLCJpYXQiOjE1MTYyMzkwMjJ9.DcS9iYP6pnATIliTA1REexXyuRuCkZMD3pugHHrB29LGY2jT6qS8evhqq-tAmzx3C0Unmu7CjglX0QAZezZM3Aa3IrKWbCVSIVjky5nO1CJ8OibC0KoK7tOUC-BrwbmeKpFX3Mjp59NfpiQD08loRNBo-g7q6vS4LR_xE78jVDb0x4ZdYboO7KJPHE40pnUEDLGT_psg_Hvtn-HFC-l76RCqxgJv3D53RwnRp0NDeAvCMEPTBfFQ931H5VFEcu9YTummdD062EAVR2KR7nYY7u4Dr2mPw4wuXDvnANjpWFuWHyw9bxB0JIiloeEAWAFjNZpT_lr_GaWrPmOk2xJzOw"
	a := MockJWTProvider()
	principal, err := a.Authenticate(token, nil)
	if err != nil {
		t.Error(err.Error())
	}
	u, ok := principal.(user)
	if !ok {
		t.Error("unexpected type of principal")
	}
	if u.GetID() != "testuser" {
		t.Error("unexpected principal")
	}
	if !u.inGroup("testgroup") {
		t.Error("User should be in testgroup")
	}
	if u.inGroup("nottestgroup") {
		t.Error("User should not be in nottestgroup")
	}
}

func TestGetInvalidUser(t *testing.T) {
	token := "notvalid"
	a := MockJWTProvider()
	_, err := a.Authenticate(token, nil)
	if err == nil {
		t.Error("Expected Error with invalid token")
	}
}

const spiffeCA = `-----BEGIN CERTIFICATE-----
MIIB5jCCAYygAwIBAgIUD/1LTTQNvk3Rp9399flLlimbgngwCgYIKoZIzj0EAwIw
UTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAkNBMRgwFgYDVQQKEw9NeSBDb21wYW55
IE5hbWUxGzAZBgNVBAMTEnVzZU9ubHlJbkRldk9yVGVzdDAeFw0xODAzMDIwMTU5
MDBaFw0yMzAzMDEwMTU5MDBaMFExCzAJBgNVBAYTAlVTMQswCQYDVQQIEwJDQTEY
MBYGA1UEChMPTXkgQ29tcGFueSBOYW1lMRswGQYDVQQDExJ1c2VPbmx5SW5EZXZP
clRlc3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARbSovOAo4ZimGBOn+tyftX
+GXShKsy2eFdvX9WfYx2NvYnw+RSM/JjRSBhUsCPXuEh/E5lhwRVfUxIlHry1CkS
o0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
jjNCAZxA5kjDK1ogrwkdziFiDgkwCgYIKoZIzj0EAwIDSAAwRQIgLXo9amyNn1Y3
qLpqrzVF7N7UQ3mxTl01MvnsqvahI08CIQCArwO8KmbPbN5XZrQ2h9zUgbsebwSG
dfOY505yMqiXig==
-----END CERTIFICATE-----`

const spiffeCertB64 = `MIIB7TCCAZOgAwIBAgIDEAAEMAoGCCqGSM49BAMCMFExCzAJBgNVBAYTAlVTMQsw
CQYDVQQIEwJDQTEYMBYGA1UEChMPTXkgQ29tcGFueSBOYW1lMRswGQYDVQQDExJ1
c2VPbmx5SW5EZXZPclRlc3QwHhcNMTgwMzAyMDI1NjEyWhcNMTkwMzAyMDI1NjEy
WjBKMQswCQYDVQQGEwJVUzELMAkGA1UECAwCQ0ExGDAWBgNVBAoMD015IENvbXBh
bnkgTmFtZTEUMBIGA1UEAwwLZXhhbXBsZS5jb20wWTATBgcqhkjOPQIBBggqhkjO
PQMBBwNCAAQQTbdQNoE5/j6mgh4HAdbgPyGbuzjpHI/x34p6qPojduUK+ifUW6Mb
bS5Zumjh31K5AmWYt4jWfU82Sb6sxPKXo2EwXzAJBgNVHRMEAjAAMAsGA1UdDwQE
AwIF4DBFBgNVHREEPjA8hhxzcGlmZmU6Ly9leGFtcGxlLmNvbS9zZXJ2aWNlggtl
eGFtcGxlLmNvbYIPd3d3LmV4YW1wbGUuY29tMAoGCCqGSM49BAMCA0gAMEUCIQDO
TaI0ltMPlPDt4XSdWJawZ4euAGXJCyoxHFs8HQK8XwIgVokWyTcajFoP0/ZfzrM5
SihfFJr39Ck4V5InJRHPPtY=`

func TestSpiffeSuccess(t *testing.T) {
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM([]byte(spiffeCA))
	a := SpiffeProvider{
		isDev: true,
		CAs:   caPool,
		time:  func() time.Time { return time.Date(2018, time.March, 22, 11, 0, 0, 0, time.UTC) },
	}
	testSpiffeAuthFlow(t, "0sspiffe://example.com/service", &a)
}

func TestSpiffeToPrincipalBadInput(t *testing.T) {
	inputs := [][]string{
		nil,
		{},
		{""},
		{"!"},
		{"foo"},
		{"spiffe:/asdf"},
		{"a", "b"},
	}

	for _, input := range inputs {
		_, err := spiffeToPrincipal(input)
		if err == nil {
			t.Errorf("Input %s should return err, but did not", input)
		}
	}
}

func testSpiffeAuthFlow(t *testing.T, authHeader string, provider Provider) {
	expected := "spiffe://example.com/service"
	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	req.Header.Add("Authorization", authHeader)
	req.RemoteAddr = "0.0.0.0:23423"
	certBytes := make([]byte, base64.StdEncoding.DecodedLen(len(spiffeCertB64)))
	n, err := base64.StdEncoding.Decode(certBytes, []byte(spiffeCertB64))
	if err != nil {
		t.Fatal(err.Error())
	}
	c, _ := x509.ParseCertificate(certBytes[:n])
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{c},
	}

	p, err := provider.Authenticate("spiffe://example.com/service", req)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !IsService(p) {
		t.Fatal("Authenticated principal is not a service")
	}
	if p.GetID() != expected {
		t.Fatalf("Service ID differs from expected result (result was %s)", p.GetID())
	}
}

func TestSpiffeFallbackSuccess(t *testing.T) {
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM([]byte(spiffeCA))
	a := SpiffeFallbackProvider{
		SpiffeProvider: SpiffeProvider{
			isDev: true,
			CAs:   caPool,
			time:  func() time.Time { return time.Date(2018, time.March, 22, 11, 0, 0, 0, time.UTC) },
		},
	}
	testSpiffeAuthFlow(t, "0sspiffe://example.com/service", &a)
}
