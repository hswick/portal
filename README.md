# Portal

A simple login application meant to be the front facing portal for internal organization apps

# Installation + Setup

Install a Go version >= 1.8

Install elm

Install postgres

```
su - postgres
createuser *username*
createdb -O *username* portal

su - *username*
psql portal
ALTER USER *username* WITH PASSWORD 'new_password';
```

Create proper config files

```
echo 'app1="supersecret"' > apps.toml
```

Create `db.toml` with proper info like this:
```
driver="postgres"
user="*username*"
password="*password*"
dbname="portal"
```

# Usage
```
go run server.go
```