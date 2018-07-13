build: pre
	docker build --rm -t xmcdev/dispatcher-srv:latest .

pre:
	./genproto.sh

debug:
	docker run -it --net=host -e DEBUG=true xmcdev/dispatcher-srv

push:
	docker push xmcdev/dispatcher-srv
