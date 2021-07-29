package authentication

type contextKey int

const (
	contextKeyTokenInfoID contextKey = iota
	contextKeyConn
)

//func (p *module) parseToken(token string) (struct{}, error) {
//	if token != p.token {
//		return struct{}{}, errors.New("Invalid session token")
//	}
//
//	// Currently this empty struct doesn't add any information
//	// to the context, but we could potentially add some custom
//	// type.
//	return struct{}{}, nil
//}
//
////grpcAuth is a middleware (interceptor) that extracts and verifies token from header
//func (p *module) grpcAuth(ctx context.Context) (context.Context, error) {
//
//	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
//	if err != nil {
//		return nil, err
//	}
//
//	tokenInfo, err := p.parseToken(token)
//	if err != nil {
//		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
//	}
//
//	// do we need this at all?
//	newCtx := context.WithValue(ctx, contextKeyTokenInfoID, tokenInfo)
//
//	return newCtx, nil
//}

//func (p *module) buildSelfSignedKeyPair() ([]byte, []byte) {
//
//	hosts := []string{"127.0.0.1", "localhost", "::1"}
//	if p.config.IpcAddress != "" {
//		hosts = append(hosts, p.config.IpcAddress)
//	}
//	_, rootCertPEM, rootKey, err := security.GenerateRootCert(hosts, 2048)
//	if err != nil {
//		return nil, nil
//	}
//
//	// PEM encode the private key
//	rootKeyPEM := pem.EncodeToMemory(&pem.Block{
//		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey),
//	})
//
//	// Create and return TLS private cert and key
//	return rootCertPEM, rootKeyPEM
//}
//
//func (p *module) initializeTLS() {
//
//	cert, key := p.buildSelfSignedKeyPair()
//	if cert == nil {
//		panic("unable to generate certificate")
//	}
//	pair, err := tls.X509KeyPair(cert, key)
//	if err != nil {
//		panic(err)
//	}
//	p.tlsKeyPair = &pair
//	p.tlsCertPool = x509.NewCertPool()
//	ok := p.tlsCertPool.AppendCertsFromPEM(cert)
//	if !ok {
//		panic("bad certs")
//	}
//
//	p.tlsAddr = p.config.IpcAddress
//}
