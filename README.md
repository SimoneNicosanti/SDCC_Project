# SDCC_Project
Questo repository contiene il codice per un sistema di storage distribuito nell'edge-cloud continuum.
Il caso d'uso per cui è stato pensato il sistema è quello di un'organizzazione che possiede diversi dispositivi (e.g. IoT) che fanno upload e download di file dal cloud.

I casi d'uso considerati per il sistema sono:
* Upload
* Download 
* Delete

Il sistema è formato dalle seguenti componenti:
* LoadBalancer, con codice in /load_balancer
* Registry, con codice in /registry
* Edge, con codice in /edge
* Client, con codice in /client

## Requisiti 
Il progetto usa **Docker** e **Docker Compose** per istanziare le diverse componenti

## Deployment
Per fare il deployment nel docker engine locale è stato preparato uno script nella directory */test* che permette di mandare in esecuzione il sistema nella sua interezza:
```bash
./system_start.sh <numeroEdgeDesiderati>
```
Si esegua il comando dalla directory */test*

## Esecuzione delle operazioni
Per eseguire le operazioni, eseguire il comando
```bash
docker exec -ti sdcc_project-client-1 ./client_run.sh
```
Ed inserire le credenziali:
* Username: test
* Password: test

Per testare il sistem,a i file da usare possono essere inseriti (per upload) o essere trovati (per download) nella directory */docker/client_files*.

Per eseguire i casi d'uso si procede come segue:
* download & fileName
* upload & fileName
* delete & fileName
Il file viene cercato in automatico nella directory */files* del container client.

## Esecuzione test
Usando gli script :
* */test/sequential_test.sh* <useCase> <edgeNum> <preliminarUpload>
* */test/parallel_test.sh* <useCase> <edgeNum> <preliminarUpload>
Si possono testare le funzionalità del sistema, prendendo i tempi di esecuzione e salvandoli in file CSV della directory "/test/Results".

```bash
/test/sequential_test.sh <useCase> <edgeNum> <preliminarUpload>
```
Dove:
* useCase può essere "up", "down_cache", "down_no_cache"
* preliminarUpload può essere "prp_up" oppure nulla, per fare upload preliminare dei file sul cloud in caso di test di download

```bash
/test/parallel_test.sh <useCase> <edgeNum> <preliminarUpload>
```
Dove:
* useCase può essere "up_all", "down_all", "mixed"
* preliminarUpload può essere "prp_up" oppure nulla, per fare upload preliminare dei file sul cloud in caso di test di download