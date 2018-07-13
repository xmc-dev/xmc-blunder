all: auth-srv xmc-core eval-srv dispatcher-srv account-srv api-srv

%-srv: pre
	docker build --rm -t xmcdev/$@:latest -f Dockerfile.$@ .

xmc-core: pre
	docker build --rm -t xmcdev/$@:latest -f Dockerfile.$@ .

pre:
	./genproto.sh
