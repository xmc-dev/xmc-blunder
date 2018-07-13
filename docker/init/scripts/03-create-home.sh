contents="$(cat <<EOF | base64
<h1>Welcome to XMC!</h1>
EOF
)"

micro call xmc.srv.core PageService.Create \
	'{"page": {"path": "/"}, "title": "Home", "contents": "'"$contents"'"}'
