# Woodpecker

Check Pod status in TWCC Kubernetes cluster

## Endpoints

### Local Test
1. Check Pod create life cycle

```bash
$ curl http://woodpecker:woodpecker@127.0.0.1:8080/selfcheck | jq .
```

2. Query allocated GPU on each node
```bash
$ curl http://woodpecker:woodpecker@127.0.0.1:8080/nodeGpuStatus | jq .
```


### TWCC

1. Check Pod create life cycle

```bash
$ curl -H 'Authorization: Basic d29vZHBlY2tlcjp3b29kcGVja2Vy' http://172.29.188.60:8080/lifeCycleCheck | jq .

{
  "CreateNamespace": "PASS",
  "CreatePod": "PASS",
  "CreateSVC": "PASS",
  "IntraConnection": "PASS",
  "InterConnection": "PASS"
}

```

2. Query allocated GPU on each node
```bash
$ curl -H 'Authorization: Basic d29vZHBlY2tlcjp3b29kcGVja2Vy' http://172.29.188.60:8080/nodeGpuStatus | jq .

{
  "Result": [
    {
      "Node": "gn0510.twcc.ai",
      "Count": 5
    },
    {
      "Node": "gn0807.twcc.ai",
      "Count": 2
    },
    {
      "Node": "gn0904.twcc.ai",
      "Count": 1
    }
  ]
} 
```
