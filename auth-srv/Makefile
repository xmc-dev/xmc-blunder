build:
	docker build --rm -t xmcdev/auth-srv:latest .

debug:
	docker run -it --net=host -e DEBUG=true xmcdev/auth-srv

push:
	docker push xmcdev/auth-srv
