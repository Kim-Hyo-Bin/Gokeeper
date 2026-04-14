import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.KeyFactory;
import java.security.PublicKey;
import java.security.Signature;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Offline Ed25519 license validation (Java 15+ {@code Ed25519}).
 *
 * Env: PUBLIC_KEY_PATH (default ./license_public.pem), LICENSE_KEY (or first arg).
 *
 * <p>Compile: {@code javac OfflineValidate.java} Run: {@code java OfflineValidate}
 */
public class OfflineValidate {

    public static void main(String[] args) throws Exception {
        String pemPath = System.getenv().getOrDefault("PUBLIC_KEY_PATH", "license_public.pem");
        String licenseKey =
                System.getenv("LICENSE_KEY") != null ? System.getenv("LICENSE_KEY") : (args.length > 0 ? args[0] : "");
        if (licenseKey.isEmpty()) {
            System.err.println("set LICENSE_KEY or pass license_key as arg");
            System.exit(1);
        }

        String[] parts = licenseKey.split("\\.");
        if (parts.length != 2) {
            System.err.println("invalid license key format");
            System.exit(1);
        }

        byte[] payload = Base64.getUrlDecoder().decode(parts[0]);
        byte[] sig = Base64.getUrlDecoder().decode(parts[1]);

        String pem = Files.readString(Path.of(pemPath), StandardCharsets.UTF_8);
        PublicKey pub = loadEd25519PublicKey(pem);

        Signature verifier = Signature.getInstance("Ed25519");
        verifier.initVerify(pub);
        verifier.update(payload);
        if (!verifier.verify(sig)) {
            System.err.println("signature verify failed");
            System.exit(1);
        }

        String json = new String(payload, StandardCharsets.UTF_8);
        String licId = extract(json, "license_id");
        long exp = Long.parseLong(extractNumber(json, "exp"));
        long now = System.currentTimeMillis() / 1000;
        if (exp != 0 && now > exp) {
            System.err.println("expired (exp=" + exp + ")");
            System.exit(1);
        }
        System.out.println("ok license_id=" + licId + " exp=" + exp);
    }

    private static PublicKey loadEd25519PublicKey(String pem) throws Exception {
        String b64 =
                pem.replace("-----BEGIN PUBLIC KEY-----", "")
                        .replace("-----END PUBLIC KEY-----", "")
                        .replaceAll("\\s", "");
        byte[] der = Base64.getDecoder().decode(b64);
        KeyFactory kf = KeyFactory.getInstance("Ed25519");
        return kf.generatePublic(new X509EncodedKeySpec(der));
    }

    private static String extract(String json, String field) {
        Pattern p = Pattern.compile("\"" + Pattern.quote(field) + "\"\\s*:\\s*\"([^\"]*)\"");
        Matcher m = p.matcher(json);
        if (!m.find()) {
            throw new IllegalArgumentException("missing field: " + field);
        }
        return m.group(1);
    }

    private static String extractNumber(String json, String field) {
        Pattern p = Pattern.compile("\"" + Pattern.quote(field) + "\"\\s*:\\s*(-?\\d+)");
        Matcher m = p.matcher(json);
        if (!m.find()) {
            return "0";
        }
        return m.group(1);
    }
}
