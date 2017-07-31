# consul_autojoin
Tool for auto joining on GCP


Goal is is to discover all consul nodes and ensure that a connection is setup between all 
datacenters. 

currently works with 2 args, project and tags

`go run main.go project tag`

Output:
```
Region: us-central1, Nodes: [10.128.0.3 10.128.0.2] 
Region: us-east4, Nodes: [10.150.0.2] 
```
