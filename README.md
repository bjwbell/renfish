# renfish
Gophish Hosting

# Launching
Execute `sudo ./launch.sh&`

# Logs
Logs are in `renfish.log`

# Gophish Container

# Pull

```
docker pull bjwbell/gophish-container
```

# Run

```
docker network create --subnet=172.19.0.0/16 gophish
docker run --net gophish --ip 172.19.0.2 bjwbell/gophish-container /gophish/gophish
```
