package io.gitfire.harness;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

class CliBridgeTest {
  @TempDir Path tmp;

  private static Path workspaceRoot() {
    return Path.of("../..").toAbsolutePath().normalize();
  }

  private static void runGit(Path dir, String... args) throws Exception {
    List<String> cmd = new java.util.ArrayList<>();
    cmd.add("git");
    for (String a : args) {
      cmd.add(a);
    }
    ProcessBuilder pb = new ProcessBuilder(cmd);
    pb.directory(dir.toFile());
    Process p = pb.start();
    boolean ok = p.waitFor(60, java.util.concurrent.TimeUnit.SECONDS);
    if (!ok) {
      p.destroyForcibly();
      throw new RuntimeException("git timeout");
    }
    if (p.exitValue() != 0) {
      String err = new String(p.getErrorStream().readAllBytes(), java.nio.charset.StandardCharsets.UTF_8);
      throw new RuntimeException("git failed: " + err);
    }
  }

  @Test
  void analyzeRepositorySeesCleanRepo() throws Exception {
    Path repo = tmp.resolve("r");
    Files.createDirectories(repo);
    runGit(repo, "init");
    Files.writeString(repo.resolve("a.txt"), "x\n");
    runGit(repo, "add", "a.txt");
    runGit(repo, "commit", "-m", "init");

    CliBridge bridge = new CliBridge(workspaceRoot());
    CliBridge.RepositoryMeta meta = bridge.analyzeRepository(repo);

    assertTrue(Files.exists(repo.resolve(".git")));
    assertFalse(meta.isDirty());
    assertTrue(meta.path().endsWith("r") || meta.path().contains("r"));
  }

  @Test
  void isDirtyDetectsUntracked() throws Exception {
    Path repo = tmp.resolve("d");
    Files.createDirectories(repo);
    runGit(repo, "init");
    Files.writeString(repo.resolve("b.txt"), "y\n");

    CliBridge bridge = new CliBridge(workspaceRoot());
    assertTrue(bridge.isDirty(repo.toString()));
  }

  @Test
  void safetySanitizeTextRemovesToken() {
    CliBridge bridge = new CliBridge(workspaceRoot());
    String token = "ghp_" + "a".repeat(36);
    String out = bridge.safetySanitizeText("pat " + token);
    assertFalse(out.contains("ghp_"));
    assertTrue(out.contains("[REDACTED]"));
  }

  @Test
  void scanRepositoriesFindsNestedRepo() throws Exception {
    Path outer = tmp.resolve("outer");
    Path inner = outer.resolve("nested").resolve("proj");
    Files.createDirectories(inner);
    runGit(inner, "init");
    Files.writeString(inner.resolve("f"), "1\n");
    runGit(inner, "add", "f");
    runGit(inner, "commit", "-m", "c");

    CliBridge bridge = new CliBridge(workspaceRoot());
    List<CliBridge.RepositoryMeta> repos =
        bridge.scanRepositories(
            new CliBridge.ScanOptions().rootPath(outer.toString()).useCache(false).maxDepth(20));

    boolean found =
        repos.stream().anyMatch(r -> r.path().equals(inner.toAbsolutePath().normalize().toString()));
    assertTrue(found);
  }
}
