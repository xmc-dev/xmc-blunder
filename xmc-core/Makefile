build: pre
	docker build --rm -t xmcdev/xmc-core:latest .

pre:
	./genproto.sh

debug:
	docker run -it --net=host -e DEBUG=true xmcdev/xmc-core

push:
	docker push xmcdev/xmc-core
