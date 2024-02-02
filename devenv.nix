{ pkgs, ... }:

{
	packages = [
		pkgs.nodejs_20
	];

	languages.go.enable = true;

	scripts.go-gen.exec = "go generate ./...";
	scripts.go-for-watch.exec = "go-gen && go $@";
	scripts.gowatch.exec = ''
		go run github.com/mitranim/gow@latest \
		-v -i .devenv -i internal/db/queries -i internal/listdb/queries -i static/tailwind.css \
		-g go-for-watch -e go,html,css,txt,sql run $@
	'';

	env.DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/membersdb?sslmode=disable";
	env.LIST_DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/listsdb?sslmode=disable";
	env.REDIS_URL = "redis://localhost:6379";
	env.LDAP_URL = "127.0.0.1:6636";
	env.SMTP_URL = "127.0.0.1:2525";
	env.LDAP_SELFSERVICE_PASSWORD = "password";
}
