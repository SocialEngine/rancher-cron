VERSION = 0.0.2
OUTPUT_FILE = docker/dist/rancher-cron

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(OUTPUT_FILE) && chmod +x $(OUTPUT_FILE)
package: build
	docker build -t socialengine/rancher-cron:$(VERSION) docker
publish: 
	docker tag socialengine/rancher-cron:$(VERSION) socialengine/rancher-cron:latest && \
	docker push socialengine/rancher-cron:$(VERSION) && \
	docker push socialengine/rancher-cron:latest
