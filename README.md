# renfish
Gophish Hosting

# Launching
Execute `sudo ./launch.sh&`

# Logs
Logs are in `renfish.log`

# SSL Certificate

## Renew
Execute `certbot renew --pre-hook "service nginx stop" --post-hook "service nginx start"`

## Add Subdomain
Execute `certbot certonly --standalone --pre-hook "service nginx stop" --post-hook "service nginx start" -d sudomain.renfish.com`

# Gophish Container

## Pull

```
docker pull bjwbell/gophish-container
```

## Run

```
docker network create --subnet=172.19.0.0/16 gophish
docker run --net gophish bjwbell/gophish-container /gophish/gophish
```
