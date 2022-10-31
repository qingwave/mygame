# Mygame
Simple Kubernetes Operator to manage 2048 game, contains crd, controller and webhook.

## Prerequisites

- Golang version: 1.16 - Get it here: https://golang.org/dl/
- kubectl - Get the latest stable version here: https://github.com/kubernetes/kubectl/releases
- kustomize - Get it here: https://github.com/kubernetes-sigs/kustomize/releases

## Build and Run operator

build binary
```bash
make build
```

build image
```bash
make docker-build
```

deploy operator in kubernetes clsuter
```bash
make deploy
```

## Test

deploy `Game` CR
```bash
kubectl apply -f config/samples/myapp_v1_game.yaml
```

get games
```bash
# get game
kubectl get game
NAME          PHASE     HOST        DESIRED   CURRENT   READY   AGE
game-sample   Running   mygame.io   1         1         1       6m

# get deploy
kubectl get deploy game-sample
NAME          READY   UP-TO-DATE   AVAILABLE   AGE
game-sample   1/1     1            1           6m

# get ingress
kubectl get ing game-sample
NAME          CLASS    HOSTS       ADDRESS        PORTS   AGE
game-sample   <none>   mygame.io   192.168.49.2   80      7m
```

play the game in your browser (must add `<Ingress Address ip> mygame.io` in your `/etc/hosts`)
