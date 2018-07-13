build: pre
	docker build --rm -t xmcdev/account-srv:latest .

pre:
	../genproto.sh

debug:
	docker run -it --net=host -e DEBUG=true xmcdev/account-srv $(OPTS)

push:
	docker push xmcdev/account-srv
