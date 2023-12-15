``` 
 _______  _______  _______  _______  _____    ___ ___ 
|     __||   _   ||   |   ||    ___||     |_ |   |   |
|    |  ||       ||       ||    ___||       | \     / 
|_______||___|___||__|_|__||_______||_______|  |___|                                       
```

Gamely is a Kubernetes operator for managing various self-hosted game servers 

## Supported Servers 

- Valheim

### Planned

- 7 Days to Die
- Minecraft
- Project Zomboid
- Ark 
- DayZ
- Rust

## Installation 

``` 
helm install gamely oci://ghcr.io/robwittman/gamely/helm/gamely
kubectl apply -f https://github.com/robwittman/gamely/releases/latest/download/server.gamely.io_valheims.yaml
```