# Pipelines-as-Code - GO API Demo

Repository for a basic Go API application to demo Pipelines as code with Openshift Pipelines.

## Build and run with Podman

### Build

```bash
podman build -t my-go-api:latest -f Containerfile
```

### Run

```bash
podman run -p 8080:8080 my-go-api:latest
```

## Test API App

### Create an item

```bash
curl -X POST -H "Content-Type: application/json" -d '{"name":"Test Item","value":111}' http://localhost:8080/items/
```

### List Items

```bash
 curl http://localhost:8080/items/ | jq .
```

### Get an item

```bash
curl http://localhost:8080/items/{id}
```

### Delete an Item

```bash
curl -X DELETE http://localhost:8080/items/{id}
```

### Update an Item

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"name":"Updated Item","value":20}' http://localhost:8080/items/{id}
```
