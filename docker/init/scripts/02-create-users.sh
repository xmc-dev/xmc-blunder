callback_url="${XMC_URL:-http://localhost:8082}"

function cred() {
	local scope="$1"
	local name="$2"
	local ret="$(micro call xmc.srv.account AccountsService.Create \
		'{"account": {"type": 1, "owner_uuid": "'"$owner_uuid"'", "callback_url": "http://localhost", "scope": "'"$scope"'", "name": "'"$name"'", "is_public": false}}')"
	test "$?" -ne 0 && exit 1

	local id="$(echo "$ret" | jq -r .client_id)"
	local secret="$(echo "$ret" | jq -r .client_secret)"

	echo "${id}:${secret}"
}

function kv_put() {
	local key="$1"
	curl --request PUT --data @- http://consul:8500/v1/kv/$key
}

owner_uuid="$(micro call xmc.srv.account AccountsService.Search '{"client_id": "root"}' | jq -r .accounts[0].uuid)"

# admin
micro call xmc.srv.account AccountsService.Create \
	'{"account": {"client_id": "admin", "name": "Teodor Romanov", "client_secret": "admin", "role_id": "admin"}}'

# normal user
micro call xmc.srv.account AccountsService.Create \
	'{"account": {"client_id": "user", "name": "Ion Ciprian", "client_secret": "user"}}'

# api bot
micro call xmc.srv.account AccountsService.Create \
	'{"account": {"type": 1, "owner_uuid": "'"$owner_uuid"'", "callback_url": "'"$callback_url"'/login", "name": "API Bot", "is_public": true}}'

cred "xmc.dispatcher/create" "XMC Core" | kv_put "xmc.srv.core/credentials"
cred "xmc.core/manage/submission xmc.eval/assign" "XMC Dispatcher" | kv_put "xmc.srv.dispatcher/credentials"
cred "xmc.core/manage/attachment xmc.dispatcher/finish" "XMC Eval" | kv_put "xmc.srv.eval/credentials"
