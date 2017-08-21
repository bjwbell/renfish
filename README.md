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
Execute `certbot run --pre-hook "service nginx stop" --post-hook "service nginx start -d renfish.com -d sudomain1.renfish.com -d subdomain2.renfish.com"`

# Gophish Container

## Pull

```
docker pull bjwbell/gophish-container
```

## Run

```
docker network create --subnet=172.19.0.0/16 gophish
docker run --net gophish --ip 172.19.0.2 bjwbell/gophish-container /gophish/gophish
```
