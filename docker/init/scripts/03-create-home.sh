contents="<h1>Welcome to XMC!</h1>"

micro call xmc.srv.core PageService.Create \
	'{"page": {"path": "/"}, "title": "Home", "contents": "'"$contents"'"}'
