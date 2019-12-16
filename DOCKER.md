readeef with docker
===================

# Docker

## Requirements

To use this docker compose file you should comply with this requirements:

* Install [Docker Desktop](https://www.docker.com/products/docker-desktop) for Windows/Mac or [Docker Engine](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-ce) for Linux  
* Install [docker-compose](https://docs.docker.com/compose/install/) (This is installed by default on Windows and Mac with Docker installation)

### Build the image
```bash
make docker-build
```

### Run the image
```bash
make docker-run
```

### Run with Docker-compose

```bash
docker-compose up -d
```
