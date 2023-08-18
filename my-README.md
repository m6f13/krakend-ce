- links: 
- server plugins: https://www.krakend.io/docs/extending/http-server-plugins/
- client plugins: https://www.krakend.io/docs/extending/http-client-plugins/

used commands:
```shell
curl -X GET -u me:pass http://localhost:8080/secure
wget --header="Authorization: Basic $(echo -n me:pass | base64)" http://localhost:8080/secure
```
