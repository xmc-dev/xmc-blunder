build:
	docker build --rm -t xmcdev/api-srv:latest .

debug:
	docker run -it --net=host -e DEBUG=true xmcdev/api-srv

push:
	docker push xmcdev/api-srv
