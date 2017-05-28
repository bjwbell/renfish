# renfish
Gophish Hosting

# Docker Container
## Build

```
docker build -t bjwbell/renfish .
```
## Push
```
docker login
docker push bjwbell/renfish

```
## Run
```
docker pull bjwbell/renfish
docker run -v /etc/letsencrypt:/etc/letsencrypt -p 9080:80 -p 9443:443 -ip 172.17.0.2 bjwbell/renfish /renfish/launch.sh
```
