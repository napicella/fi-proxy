# fi-proxy PoC
Failure injector proxy.  
Transparent egress proxy which inject latency to your dependencies to test how the system reacts to increased 
latency/failures.  
This is a PoC, it needs some work to be used:
- failures configuration requires changing the code
- there are no tests in place

## Usage
For now the failures are encoded directly in the main:  
You need to change the endpoint that should cause latency/failures in the main.

After that, rebuild and run it:
```bash
go build
./fi-proxy --proto http
``` 

Start the service under test, using the proxy.
This depends on the language used for the service, i.e. for java you need to start it with:
```bash
## Proxy http/s to localhost:8888
JAVA_FLAGS="-Dhttp.proxyHost=localhost -Dhttp.proxyPort=8888 -Dhttps.proxyHost=localhost -Dhttps.proxyPort=8888 -Dhttp.nonProxyHosts="
## Start your service
java ${JAVA_FLAGS} -jar your-service.jar 
```

### Reference
Transparent proxy code is inspired by [@mlowicki's blog post](https://medium.com/@mlowicki/http-s-proxy-in-golang-in-less-than-100-lines-of-code-6a51c2f2c38c)
