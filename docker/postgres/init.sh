set -e
set -u

alias psql="psql -v ON_ERROR_STOP=1 --username $POSTGRES_USER"

psql <<-EOF
	CREATE USER dev;
	ALTER USER dev WITH ENCRYPTED PASSWORD 'pass';
EOF

for d in core account dispatcher; do
	psql <<-EOF
		CREATE DATABASE $d;
		GRANT ALL ON DATABASE $d TO dev;
	EOF

	psql -d $d <<-EOF
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE EXTENSION IF NOT EXISTS "ltree";
	EOF
done
