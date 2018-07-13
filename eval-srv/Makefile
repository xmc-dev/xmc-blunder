build: pre
	docker build --rm -t xmcdev/eval-srv:latest .

pre:
	./genproto.sh

debug:
	docker run -it --net=host -e DEBUG=true xmcdev/eval-srv

push:
	docker push xmcdev/eval-srv
