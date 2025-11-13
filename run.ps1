$env:CGO_ENABLED = "1"
$env:APP_ENV = "development"
$env:PORT = "8080"
$env:JWT_SECRET = "devsecret"
go run ./cmd/app

