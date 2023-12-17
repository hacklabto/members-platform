{ pkgs, ... }:

{
	packages = [
		pkgs.nodejs_20
	];

	languages.go.enable = true;

	scripts.go-gen.exec = "go generate ./...";
	scripts.go-for-watch.exec = "go-gen && go $@";
	scripts.watch-web.exec = ''
		go run github.com/mitranim/gow@latest \
		-v -i .devenv -i internal/db/queries -i static/tailwind.css \
		-g go-for-watch -e go,html,css,txt run ./cmd/web
	'';

	env.DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/membersdb?sslmode=disable";
	env.LDAP_URL = "127.0.0.1:6636";
	env.PASSWD_RESET_HASHER_SECRET = "devsecret";
}
