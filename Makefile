all: install clean build

test: test-go

test-go:
	(cd ./lambda/go && go test ./...)

install:
	npm install

clean:
	npm run clean

format:
	npm run format-fix
	(cd ./lambda/go && gofmt -w -s ./)

build:
	npm run build

synth: test
	sam-beta-cdk build

deploy: clean build synth
	cdk deploy -a .aws-sam/build

bootstrap: install
	./scripts/bootstrap-config.sh
	cdk bootstrap

export-api-url:
	@echo "export API_URL=$(shell ./scripts/get-api-url.sh)"

verify-api:
	./scripts-verify-api.sh

put-test-data:
	./scripts/populate-table.sh

reset-table:
	./scripts/reset-table.sh

throttle-table:
	./scripts/throttle-table.sh

clean-workshop-data:
	./scripts/reset-table.sh

#------------------------------
# Tail Handler logs
#------------------------------
tail-logs-GetPodcast:
	./scripts/tail-logs.sh GetPodcast

tail-logs-ListPodcasts:
	./scripts/tail-logs.sh ListPodcasts

tail-logs-UploadPodcast:
	./scripts/tail-logs.sh UploadPodcast

tail-logs-PlayPodcast:
	./scripts/tail-logs.sh PlayPodcast

FINAL_BRANCH ?= refs/remotes/origin/final
MIDDLE_BRANCH ?= refs/remotes/origin/middle

#------------------------------
# Check workshop handler diff
#------------------------------
ws-check-diff-GetPodcast:
	./scripts/ws-check.sh --diff "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh GetPodcast)"

ws-check-diff-ListPodcasts-part1:
	./scripts/ws-check.sh --diff "${MIDDLE_BRANCH}" "$(shell ./scripts/ws-handler-file.sh ListPodcasts)"

ws-check-diff-PlayPodcast-part1:
	./scripts/ws-check.sh --diff "${MIDDLE_BRANCH}" "$(shell ./scripts/ws-handler-file.sh PlayPodcast)"

ws-check-diff-ListPodcasts-part2:
	./scripts/ws-check.sh --diff "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh ListPodcasts)"

ws-check-diff-PlayPodcast-part2:
	./scripts/ws-check.sh --diff "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh PlayPodcast)"

ws-check-diff-UploadPodcast:
	./scripts/ws-check.sh --diff "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh UploadPodcast)"

#------------------------------
# Check workshop handler print
#------------------------------
ws-check-print-GetPodcast:
	./scripts/ws-check.sh --print "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh GetPodcast)"

ws-check-print-ListPodcasts-part1:
	./scripts/ws-check.sh --print "${MIDDLE_BRANCH}" "$(shell ./scripts/ws-handler-file.sh ListPodcasts)"

ws-check-print-PlayPodcast-part1:
	./scripts/ws-check.sh --print "${MIDDLE_BRANCH}" "$(shell ./scripts/ws-handler-file.sh PlayPodcast)"

ws-check-print-ListPodcasts-part2:
	./scripts/ws-check.sh --print "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh ListPodcasts)"

ws-check-print-PlayPodcast-part2:
	./scripts/ws-check.sh --print "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh PlayPodcast)"

ws-check-print-UploadPodcast:
	./scripts/ws-check.sh --print "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh UploadPodcast)"

#------------------------------
# Check workshop handler copy
#------------------------------
ws-check-copy-GetPodcast:
	./scripts/ws-check.sh --copy "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh GetPodcast)"

ws-check-copy-ListPodcasts-part1:
	./scripts/ws-check.sh --copy "${MIDDLE_BRANCH}" "$(shell ./scripts/ws-handler-file.sh ListPodcasts)"

ws-check-copy-PlayPodcast-part1:
	./scripts/ws-check.sh --copy "${MIDDLE_BRANCH}" "$(shell ./scripts/ws-handler-file.sh PlayPodcast)"

ws-check-copy-ListPodcasts-part2:
	./scripts/ws-check.sh --copy "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh ListPodcasts)"

ws-check-copy-PlayPodcast-part2:
	./scripts/ws-check.sh --copy "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh PlayPodcast)"

ws-check-copy-UploadPodcast:
	./scripts/ws-check.sh --copy "${FINAL_BRANCH}" "$(shell ./scripts/ws-handler-file.sh UploadPodcast)"
