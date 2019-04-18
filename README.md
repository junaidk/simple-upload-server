


#### Run image
```bash
docker run -it -p 1323:1323  -v $(pwd)/db:/app/db -v $(pwd)/data:/app/data filesrv
```


#### ENV
    - DATA_DIR path
    - DB_DIR path
    - IP_REST true/false
    
    
#### Get csv data

```bash
docker run -it -p 1323:1323  -v $(pwd)/db:/app/db -v $(pwd)/data:/app/data filesrv /app/uploadsrv csv
```

output will be in `data dir`