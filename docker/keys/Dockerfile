FROM consul

COPY pubkey.pem /keys/pubkey.pem
COPY privkey.pem /keys/privkey.pem

CMD consul kv put -http-addr=consul:8500 xmc.srv.auth/pubkey "$(cat /keys/pubkey.pem)" && consul kv put -http-addr=consul:8500 xmc.srv.auth/privkey "$(cat /keys/privkey.pem)"
