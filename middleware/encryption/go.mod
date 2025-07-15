module schneider.vip/hybridbuffer/middleware/encryption

go 1.23.0

toolchain go1.24.0

require (
	github.com/minio/sio v0.4.1
	github.com/pkg/errors v0.9.1
	schneider.vip/hybridbuffer/middleware v1.0.2
)

require (
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
)

replace schneider.vip/hybridbuffer/middleware => ../
