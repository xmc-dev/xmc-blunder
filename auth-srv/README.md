# XMC's Authorization Server (auth-srv)

`auth-srv` is the authentication component of XMC. It issues tokens that human users and
automated services can use to interact with XMC Platform, following the [OAuth
2.0][oauth2] spec.

## Programs required

* Go 1.8+
* Consul
* Redis
* [account-srv][account-srv]

[oauth2]: https://oauth.net/2/
[account-srv]: https://github.com/xmc-dev/account-srv
