Envoi d'une DPAE vers l'URSSAF
==============================

Projet à titre d'exemple et recherche.

Installation

`go get github.com/flibustenet/dpae`

Installer l'exécutable `dpae`

`go install`

Ou compiler sur place

`go build -o run_dpae`

Utilisation
-----------

Créer un fichier `dpae_test.json` en suivant l'exemple `dpae_sample.json`

Tester ce fichier avec l'exécutable compilé précédement.

`./run_dpae dpae_test.json`

Il va renvoyer `idflux` et le certificat.

`./run_dpae -v dpae_test.json`

Idem mais affiche également le fichier xml envoyé.

Librairie
---------

Il est également possible de l'utiliser sous forme d'une librairie en Go
en prenant exemple sur le fichier `main.go`.

Soit en lisant un fichier json soit en construisant directement le
struct `dpae.Dpae`

Tests
-----

Pour lancer les tests fonctionnels sur le fichier `dpae_test.json` :

```
$ export DPAE_TEST_JSON=dpae_test.json
$ go test -v ./... -count=1
```



