import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * HTTP API: issue (7-day), get, verify, revoke. Java 17+ (java.net.http).
 *
 * Env: BASE_URL (default http://localhost:8080)
 *
 * <p>Compile: {@code javac ApiExample.java} Run: {@code java ApiExample}
 */
public class ApiExample {

    public static void main(String[] args) throws Exception {
        String base = System.getenv().getOrDefault("BASE_URL", "http://localhost:8080").replaceAll("/$", "");
        HttpClient client = HttpClient.newHttpClient();

        String exp = Instant.now().plus(7, ChronoUnit.DAYS).toString();
        String issueBody = "{\"expires_at\":\"" + exp + "\"}";
        HttpRequest issueReq =
                HttpRequest.newBuilder()
                        .uri(URI.create(base + "/v1/licenses"))
                        .header("Content-Type", "application/json")
                        .POST(HttpRequest.BodyPublishers.ofString(issueBody))
                        .build();
        HttpResponse<String> issueRes = client.send(issueReq, HttpResponse.BodyHandlers.ofString());
        if (issueRes.statusCode() != 201) {
            throw new IllegalStateException("issue: " + issueRes.statusCode() + " " + issueRes.body());
        }
        String body = issueRes.body();
        String id = extract(body, "id");
        String licenseKey = extract(body, "license_key");
        System.out.println("issued id: " + id);

        HttpRequest getReq = HttpRequest.newBuilder().uri(URI.create(base + "/v1/licenses/" + id)).GET().build();
        HttpResponse<String> getRes = client.send(getReq, HttpResponse.BodyHandlers.ofString());
        System.out.println("get: " + getRes.statusCode());

        String verifyBody = "{\"license_key\":\"" + escapeJson(licenseKey) + "\"}";
        HttpRequest verifyReq =
                HttpRequest.newBuilder()
                        .uri(URI.create(base + "/v1/licenses/verify"))
                        .header("Content-Type", "application/json")
                        .POST(HttpRequest.BodyPublishers.ofString(verifyBody))
                        .build();
        HttpResponse<String> verifyRes = client.send(verifyReq, HttpResponse.BodyHandlers.ofString());
        System.out.println("verify: " + verifyRes.body());

        HttpRequest revokeReq =
                HttpRequest.newBuilder()
                        .uri(URI.create(base + "/v1/licenses/" + id + "/revoke"))
                        .POST(HttpRequest.BodyPublishers.noBody())
                        .build();
        HttpResponse<String> revokeRes = client.send(revokeReq, HttpResponse.BodyHandlers.ofString());
        System.out.println("revoke: " + revokeRes.statusCode() + " " + revokeRes.body());
    }

    private static String extract(String json, String field) {
        Pattern p = Pattern.compile("\"" + Pattern.quote(field) + "\"\\s*:\\s*\"([^\"]*)\"");
        Matcher m = p.matcher(json);
        if (!m.find()) {
            throw new IllegalArgumentException("missing field: " + field);
        }
        return m.group(1);
    }

    /** Minimal escape for license_key inside JSON string (backslash and quote). */
    private static String escapeJson(String s) {
        return s.replace("\\", "\\\\").replace("\"", "\\\"");
    }
}
