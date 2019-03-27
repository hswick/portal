psql -d portal -a -f sql/test.sql
go run server.go middleware.go
