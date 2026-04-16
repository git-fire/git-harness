package io.gitfire.harness;

import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;

public final class CliBridge {
  private static final Gson GSON = new Gson();

  public record RemoteEntry(String name, String url) {}

  public record RepositoryMeta(
      String path,
      String name,
      List<RemoteEntry> remotes,
      List<String> branches,
      boolean isDirty,
      String lastModified,
      boolean selected,
      String mode) {}

  public record WorktreeInfo(String path, String branch, String head, boolean isMain) {}

  static final class CliResult {
    final String stdout;
    final String stderr;
    final int code;

    CliResult(String stdout, String stderr, int code) {
      this.stdout = stdout;
      this.stderr = stderr;
      this.code = code;
    }
  }

  public static final class ScanOptions {
    private String rootPath = ".";
    private final List<String> exclude = new ArrayList<>();
    private int maxDepth;
    private Boolean useCache;
    private String cacheFile = "";
    private String cacheTtl = "";
    private int workers;
    private final Map<String, Boolean> knownPaths = new LinkedHashMap<>();
    private boolean disableScan;

    public ScanOptions rootPath(String v) {
      this.rootPath = v;
      return this;
    }

    public ScanOptions exclude(List<String> v) {
      exclude.clear();
      if (v != null) {
        exclude.addAll(v);
      }
      return this;
    }

    public ScanOptions maxDepth(int v) {
      this.maxDepth = v;
      return this;
    }

    public ScanOptions useCache(boolean v) {
      this.useCache = v;
      return this;
    }

    public ScanOptions cacheFile(String v) {
      this.cacheFile = v == null ? "" : v;
      return this;
    }

    public ScanOptions cacheTtl(String v) {
      this.cacheTtl = v == null ? "" : v;
      return this;
    }

    public ScanOptions workers(int v) {
      this.workers = v;
      return this;
    }

    public ScanOptions knownPath(String path, boolean rescanSubmodules) {
      knownPaths.put(path, rescanSubmodules);
      return this;
    }

    public ScanOptions disableScan(boolean v) {
      this.disableScan = v;
      return this;
    }

    private JsonObject toJson() {
      JsonObject o = new JsonObject();
      o.addProperty("rootPath", rootPath);
      o.addProperty("disableScan", disableScan);
      if (!exclude.isEmpty()) {
        JsonArray arr = new JsonArray();
        for (String s : exclude) {
          arr.add(s);
        }
        o.add("exclude", arr);
      }
      if (maxDepth > 0) {
        o.addProperty("maxDepth", maxDepth);
      }
      if (useCache != null) {
        o.addProperty("useCache", useCache);
      }
      if (!cacheFile.isEmpty()) {
        o.addProperty("cacheFile", cacheFile);
      }
      if (!cacheTtl.isEmpty()) {
        o.addProperty("cacheTTL", cacheTtl);
      }
      if (workers > 0) {
        o.addProperty("workers", workers);
      }
      if (!knownPaths.isEmpty()) {
        JsonObject kp = new JsonObject();
        for (Map.Entry<String, Boolean> e : knownPaths.entrySet()) {
          kp.addProperty(e.getKey(), e.getValue());
        }
        o.add("knownPaths", kp);
      }
      return o;
    }
  }

  private final List<String> cliCommandArgs;
  private final Path workspaceRoot;
  private final java.util.function.Function<String, CliResult> cliInvoker;

  public CliBridge(Path workspaceRoot) {
    this(workspaceRoot, defaultCliCommandArgs(workspaceRoot), null);
  }

  public CliBridge() {
    this(detectWorkspaceRoot());
  }

  CliBridge(Path workspaceRoot, java.util.function.Function<String, CliResult> cliInvoker) {
    this(workspaceRoot, defaultCliCommandArgs(workspaceRoot), cliInvoker);
  }

  CliBridge(
      Path workspaceRoot,
      List<String> cliCommandArgs,
      java.util.function.Function<String, CliResult> cliInvoker) {
    this.workspaceRoot = workspaceRoot;
    this.cliCommandArgs = List.copyOf(cliCommandArgs);
    this.cliInvoker = cliInvoker;
  }

  private static List<String> defaultCliCommandArgs(Path workspaceRoot) {
    String configuredCli = System.getenv("GIT_HARNESS_CLI");
    if (configuredCli != null && !configuredCli.isBlank()) {
      Path cliPath = Paths.get(configuredCli);
      if (!cliPath.isAbsolute()) {
        cliPath = workspaceRoot.resolve(cliPath).normalize();
      }
      return List.of(cliPath.toString());
    }
    return List.of("go", "run", "./cmd/git-harness-cli");
  }

  private static Path detectWorkspaceRoot() {
    Path current = Path.of(System.getProperty("user.dir")).toAbsolutePath().normalize();
    Path probe = current;
    while (probe != null) {
      if (Files.exists(probe.resolve("go.mod"))
          && Files.exists(probe.resolve("cmd/git-harness-cli/main.go"))) {
        return probe;
      }
      probe = probe.getParent();
    }
    return current;
  }

  public List<RepositoryMeta> scanRepositories(ScanOptions options) {
    JsonObject req = new JsonObject();
    req.addProperty("op", "scan_repositories");
    req.add("scanOptions", options.toJson());
    JsonObject res = invokeObject(req);
    List<RepositoryMeta> out = new ArrayList<>();
    if (!res.has("repositories")) {
      return out;
    }
    for (JsonElement el : res.getAsJsonArray("repositories")) {
      out.add(parseRepository(el.getAsJsonObject()));
    }
    return out;
  }

  public RepositoryMeta analyzeRepository(Path repoPath) {
    JsonObject req = new JsonObject();
    req.addProperty("op", "analyze_repository");
    req.addProperty("repoPath", repoPath.toString());
    JsonObject res = invokeObject(req);
    return parseRepository(res.getAsJsonObject("repository"));
  }

  public boolean isDirty(String repoPath) {
    JsonObject req = new JsonObject();
    req.addProperty("op", "git_is_dirty");
    req.addProperty("repoPath", repoPath);
    return invokeObject(req).get("dirty").getAsBoolean();
  }

  public String getCurrentBranch(String repoPath) {
    JsonObject req = new JsonObject();
    req.addProperty("op", "git_get_current_branch");
    req.addProperty("repoPath", repoPath);
    return invokeObject(req).get("branch").getAsString();
  }

  public String getCommitSHA(String repoPath, String ref) {
    JsonObject req = new JsonObject();
    req.addProperty("op", "git_get_commit_sha");
    req.addProperty("repoPath", repoPath);
    req.addProperty("ref", ref);
    return invokeObject(req).get("sha").getAsString();
  }

  public String safetySanitizeText(String text) {
    JsonObject req = new JsonObject();
    req.addProperty("op", "safety_sanitize_text");
    req.addProperty("text", text == null ? "" : text);
    JsonObject res = invokeObject(req);
    return res.has("text") && !res.get("text").isJsonNull() ? res.get("text").getAsString() : "";
  }

  public String safetySecurityNotice() {
    JsonObject req = new JsonObject();
    req.addProperty("op", "safety_security_notice");
    JsonObject res = invokeObject(req);
    return res.has("notice") && !res.get("notice").isJsonNull()
        ? res.get("notice").getAsString()
        : "";
  }

  public List<WorktreeInfo> listWorktrees(String repoPath) {
    JsonObject req = new JsonObject();
    req.addProperty("op", "git_list_worktrees");
    req.addProperty("repoPath", repoPath);
    JsonObject res = invokeObject(req);
    List<WorktreeInfo> out = new ArrayList<>();
    if (!res.has("worktrees")) {
      return out;
    }
    for (JsonElement el : res.getAsJsonArray("worktrees")) {
      JsonObject w = el.getAsJsonObject();
      out.add(
          new WorktreeInfo(
              w.get("path").getAsString(),
              w.get("branch").getAsString(),
              w.get("head").getAsString(),
              w.get("isMain").getAsBoolean()));
    }
    return out;
  }

  private JsonObject invokeObject(JsonObject request) {
    String raw = invokeRaw(GSON.toJson(request));
    JsonObject obj = JsonParser.parseString(raw).getAsJsonObject();
    if (!obj.has("ok") || !obj.get("ok").getAsBoolean()) {
      String err = obj.has("error") ? obj.get("error").getAsString() : "unknown error";
      throw new RuntimeException(err);
    }
    return obj;
  }

  private String invokeRaw(String payload) {
    CliResult result = runCli(payload);
    String stdout = result.stdout == null ? "" : result.stdout.trim();
    String stderr = result.stderr == null ? "" : result.stderr.trim();
    if (stdout.isBlank() && result.code != 0) {
      throw new RuntimeException("CLI failed with code " + result.code + ": " + stderr);
    }
    if (stdout.isBlank()) {
      throw new RuntimeException("CLI returned empty response");
    }
    JsonObject head = JsonParser.parseString(stdout).getAsJsonObject();
    if (result.code != 0 || !head.has("ok") || !head.get("ok").getAsBoolean()) {
      String err = head.has("error") ? head.get("error").getAsString() : stderr;
      throw new RuntimeException(err.isEmpty() ? "CLI failed with code " + result.code : err);
    }
    return stdout;
  }

  private CliResult runCli(String payload) {
    if (cliInvoker != null) {
      return cliInvoker.apply(payload);
    }
    Process process = null;
    final Process[] processRef = new Process[1];
    ExecutorService streamReaderPool = null;
    Future<String> stdoutFuture = null;
    Future<String> stderrFuture = null;
    try {
      ProcessBuilder pb = new ProcessBuilder(cliCommandArgs);
      pb.directory(workspaceRoot.toFile());
      process = pb.start();
      processRef[0] = process;
      streamReaderPool = Executors.newFixedThreadPool(2);
      stdoutFuture =
          streamReaderPool.submit(
              () ->
                  new String(
                      processRef[0].getInputStream().readAllBytes(), StandardCharsets.UTF_8));
      stderrFuture =
          streamReaderPool.submit(
              () ->
                  new String(
                      processRef[0].getErrorStream().readAllBytes(), StandardCharsets.UTF_8));
      process.getOutputStream().write(payload.getBytes(StandardCharsets.UTF_8));
      process.getOutputStream().close();
      boolean completed = process.waitFor(120, TimeUnit.SECONDS);
      if (!completed) {
        throw new RuntimeException("CLI process timed out after 120 seconds");
      }
      int code = process.exitValue();
      String out = stdoutFuture.get();
      String err = stderrFuture.get();
      return new CliResult(out, err, code);
    } catch (IOException ex) {
      throw new RuntimeException("failed to invoke CLI", ex);
    } catch (ExecutionException ex) {
      throw new RuntimeException("failed to read CLI output", ex);
    } catch (InterruptedException ex) {
      Thread.currentThread().interrupt();
      throw new RuntimeException("interrupted while invoking CLI", ex);
    } finally {
      if (process != null && process.isAlive()) {
        process.destroyForcibly();
      }
      if (stdoutFuture != null) {
        stdoutFuture.cancel(true);
      }
      if (stderrFuture != null) {
        stderrFuture.cancel(true);
      }
      if (streamReaderPool != null) {
        streamReaderPool.shutdownNow();
        try {
          streamReaderPool.awaitTermination(5, TimeUnit.SECONDS);
        } catch (InterruptedException ex) {
          Thread.currentThread().interrupt();
        }
      }
    }
  }

  private static RepositoryMeta parseRepository(JsonObject o) {
    List<RemoteEntry> remotes = new ArrayList<>();
    if (o.has("remotes") && !o.get("remotes").isJsonNull()) {
      for (JsonElement el : o.getAsJsonArray("remotes")) {
        JsonObject r = el.getAsJsonObject();
        remotes.add(new RemoteEntry(r.get("name").getAsString(), r.get("url").getAsString()));
      }
    }
    List<String> branches = new ArrayList<>();
    if (o.has("branches") && !o.get("branches").isJsonNull()) {
      for (JsonElement el : o.getAsJsonArray("branches")) {
        branches.add(el.getAsString());
      }
    }
    String lastMod = "";
    if (o.has("lastModified") && !o.get("lastModified").isJsonNull()) {
      lastMod = o.get("lastModified").getAsString();
    }
    return new RepositoryMeta(
        o.get("path").getAsString(),
        o.get("name").getAsString(),
        remotes,
        branches,
        o.get("isDirty").getAsBoolean(),
        lastMod,
        o.get("selected").getAsBoolean(),
        o.get("mode").getAsString());
  }
}
