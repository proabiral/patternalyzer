### Usage 

```
% go run main.go /tmp/example.com.txt | sort | uniq -c | sort -n | awk '{print $2}' | tail -n 10
```
