install-dep:
	@dep ensure

build:
	@go build
	sudo mv chrn /usr/local/bin/chrn
