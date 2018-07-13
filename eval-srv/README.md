# XMC's User Submission Evaluation Service (eval-srv)

`eval-srv` receives evaluation jobs from the [dispatcher][dispatcher-srv] and
compiles, runs and grades the specified code on certain test cases (the evaluation procedure).

## Required programs:

* Go 1.8+
* Consul
* MySQL / PostgreSQL (might be replaced with Redis)
* Everything that [isowrap][isowrap] requires

[dispatcher-srv]: https://github.com/xmc-dev/dispatcher-srv
[isowrap]: https://github.com/xmc-dev/isowrap
