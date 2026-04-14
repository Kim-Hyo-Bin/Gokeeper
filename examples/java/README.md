# Java examples

Requires **JDK 17+** (for `HttpClient`) and **JDK 15+** crypto for **Ed25519** offline validation.

```bash
javac ApiExample.java OfflineValidate.java
java ApiExample
java OfflineValidate   # LICENSE_KEY=... or pass key as first argument
```

`ApiExample` uses regex to parse JSON so the demo stays dependency-free; production code should use Jackson, Gson, etc.
