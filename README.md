# Portal

A simple login application meant to be the front facing portal for internal organization apps

# Installation + Setup

Install a Go version >= 1.8

Install elm + uglifyjs
```bash
npm install -g elm uglify-js
```

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
```bash
# Only need to do this once
./scripts/compile_login.sh
./scripts/compile_welcome.sh

# Run development environment
./scripts/dev.sh

# Clean database
./scripts/clean.sh
```

# Test
```bash
./scripts/test.sh
```