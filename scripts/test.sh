psql -d portal -a -f sql/test.sql
go test -v
psql -d portal -a -f sql/clean.sql
