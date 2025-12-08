.PHONY:all
all: test

.PHONY:clean
clean:
	rm reaction-pics
	rm server/static/gen/*

.PHONY:bins
bins:
	go build .

.PHONY:web
webpack:
	pnpm run build:dev

.PHONY:test
test: bins
	./bin/test.sh
	pnpm test
	git ls-files | grep -e \.sh$ | xargs shellcheck --exclude=SC1091

.PHONY:serve
serve: webpack bins
	./reaction-pics

.PHONY:lsremote
lsremote:
	bin/rclone --config=bin/rclone.conf ls backblaze:

.PHONY:backupremote
backupremote:
	bin/rclone --config=bin/rclone.conf sync backblaze: ~/storage/backup-github/reaction-pics-images
