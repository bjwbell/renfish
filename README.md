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
docker run -v /etc/letsencrypt/live/renfish.com:/etc/letsencrypt/live/renfish.com -p 80:80 -p 443:443 bjwbell/renfish /renfish/renfish
```
