# XMC - XMC Manages Contests

_Instanta publica de test este https://x.xmc.fun_

## Scopul

XMC este o platforma online pentru organizarea si desfasurarea concursurilor
de natura algoritmica si pentru pregatirea utilizatorilor cu ajutorul
problemelor de informatica din "arhiva de probleme", prin intermediul unui
evaluator automat care compileaza, executa si evalueaza solutiile
utilizatorilor fara interventie umana. Conceptul nu este nou, el fiind
intalnit in platforme precum [infoarena][infoarena],
[codeforces][codeforces], [csacademy][csacademy] etc.

[infoarena]: https://infoarena.ro/
[codeforces]: http://codeforces.com/
[csacademy]: https://csacademy.com/

XMC se deosebeste prin scopul lui de a nu fi prezent doar intr-o singura
instanta oficiala, precum exemplele de mai sus, ci pentru a fi usor de
instalat pentru oricine doreste sa organizeze un concurs _de orice marime_ sau
care vrea sa puna la dispozitie materiale de pregatire pentru elevi sau
pentru un cerc de informatica. Materialele consta in probleme de informatica,
lectii, ghiduri.

XMC se mai deosebeste si prin arhitectura si tehnologiile moderne pe care se
bazeaza. XMC este compus dintr-o suita de programe separate (_servicii_) care
comunica intre ele si care pot fi executate in mai multe instante pentru a
creste scalabilitatea si redundanta, asemanator _microserviciilor_.

## Concepte de baza

O instanta XMC este in primul rand un wiki simplu. Utilizatorii cu drepturi au
permisiunea de a creea pagini statice sau dinamice, scrise intr-un limbaj
derivat din markdown numit _XMCML_ (XMC Markup Language). Fiecare pagina poate
sa aiba si _atasamente_ asociate, adica fisiere care sunt asociate cu pagina cum
ar fi poze, documente etc.

Majoritatea obiectelor din sistemul XMC se bazeaza mai mult sau mai putin pe
pagini de wiki si atasamente. Atasamentele pot fi asociate direct obiectelor in
loc sa fie asociate paginilor.

### Lista de probleme (_Task List_)

Lista de probleme este o colectie de probleme la care utilizatorii pot _participa_,
rezolvand probleme pentru a fi situati intr-un clasament. Lista de probleme are si o
pagina de wiki de prezentare. O lista de probleme poate fi activa doar pentru un
anume interval de timp, dupa care trimiterea de solutii la problemele din lista
este oprita. Aceasta functionalitate este utila pentru concursuri si teme.

### Problema (_Task_)

O problema reprezinta o sarcina pe care utilizatorul trebuie sa o
efectueze pentru a primi puncte. Fiecare sarcina are un _enunt_, un _evaluator_
si _restrictii_. O problema poate fi asociata unei singure liste de probleme. Enuntul este
o pagina de wiki.

### Submission-ul (_Submission_)

Un submission este o incercare de rezolvare a problemei. Ea provine de
utilizator. Submission-ul este invalid daca nu respecta restrictiile problemei.
Scorul submission-ului este determinat de catre evaluatorul problemei asociate.
Codul sursa al submission-ului este un atasament.

### Setul de date (_Dataset_)

Un set de date contine testele, restrictiile si evaluatorul unei probleme.
Astfel, o problema este compusa de fapt din enunt si set de date. Un set de date
poate fi asociat cu mai multe probleme. Fiecare componenta a fiecarui test
(input si output) este un atasament.

## Instalarea sistemului

### Calea rapida

Intrati in directorul `docker` din acest proiect si porniti XMC folosing
docker-compose:

```bash
docker-compose up
```

In aproximativ un minut, caci trebuie sa descarce si sa compileze imaginile,
backend-ul XMC va fi gata.

Pentru frontend:

1. Executa intr-un terminal comanda:

```bash
micro call xmc.srv.account AccountsService.Search
```

Ultimul rezultat o sa contina un camp `client_id` cu o valoare lunga. Copiaza
aceasta valoare in clipboard.

2. Creeaza un symlink la config-ul de test in directorul `src` al frontendului:

```bash
cd ~/src/xmc/frontend # inlocuieste cum trebuie, evident :)
cd src
ln -s config.dev.js config.js
```


3. Apoi, executa `yarn dev`.

Pe sistemul host la http://localhost:8082 ar trebui sa apara interfata.

## Tehnologia si arhitectura

### Backend

Lista tehnologiilor folosite in backend:

* Go
* Postgresql
* Redis
* [Go-Micro](https://github.com/micro/go-micro) - Platforma de RPC pentru
	sisteme distribuite
* [Consul](https://consul.io) - Service discovery si sincronizare de setari
* [isolate](https://github.com/ioi/isolate) - Sandbox bazat pe Linux Containers
	* Cu ajutorul lui [isowrap](https://github.com/xmc-dev/isowrap)
* [Traefik](https://traefik.io) - Reverse proxy pentru componentele web
* [S3](https://aws.amazon.com/s3/) - Serviciu de storage de la Amazon.
	* In development si pe serverul de test folosim [Minio](https://www.minio.io/), un server de storage cu API compatibil S3.

#### Componente

* [xmc-core](https://github.com/xmc-dev/xmc-core) - Componenta principala, se
	ocupa de administrarea obiectelor (task, dataset, page etc)
* [account-srv](https://github.com/xmc-dev/account-srv) - Gestionarea conturilor
	si a sesiunilor. Conturile pot fi de utilizator sau de serviciu (roboti).
* [auth-srv](https://github.com/xmc-dev/auth-srv) - Server de autorizare care
	implementeaza framework-ul OAuth2. Bazat pe [osin](https://github.com/RangelReale/osin). Token-urile sunt de forma [JSON Web Tokens](https://jwt.io) si sunt validate folosind o pereche de chei RSA.
* [eval-srv](https://github.com/xmc-dev/eval-srv) - Primeste "job-uri" de
	evaluare de submisii de la dispecer si le evalueaza intr-un sandbox.
	Rezultatul este trimis inapoi la xmc-core.
* [dispatcher-srv](https://github.com/xmc-dev/dispatcher-srv) - Dispecerul de
	job-uri. Primeste job-uri de la xmc-core si le atribuie serverelor de
	evaluare libere.
* [api-srv](https://github.com/xmc-dev/api-srv) - Server de API REST. Expune
	informatii de la xmc-core prin JSON.

### Frontend

Repo: https://github.com/xmc-dev/web

Tehnologii:

* HTML5, CSS3, JS (ES2016), [fetch](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API)
* JSX, Babel, Webpack
* [Preact](https://github.com/developit/preact), Redux
* [Monaco Editor](https://github.com/Microsoft/monaco-editor)

#### Design:

* [Semantic UI React](https://react.semantic-ui.com/)

## Cod scris de altcineva

* [registry] este un fork al [acestui repo](https://github.com/DimShadoWWW/go-micro-consul-traefik) cu modificari aduse de noi pentru integrarea cu consul si cu versiuni noi Micro.

[registry]: https://github.com/xmc-dev/registry
