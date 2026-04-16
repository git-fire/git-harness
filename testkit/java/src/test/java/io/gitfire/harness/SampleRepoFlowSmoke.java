package io.gitfire.harness;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

/** Runnable sample that exercises scan + analyze + SHA round-trip with a real git repo. */
class SampleRepoFlowSmoke {
  @TempDir Path tmp;

  private static Path workspaceRoot() {
    return Path.of("../..").toAbsolutePath().normalize();
  }

  private static void runGit(Path dir, String... args) throws Exception {
    runGitStdout(dir, args);
  }

  private static String runGitStdout(Path dir, String... args) throws Exception {
    List<String> cmd = new ArrayList<>();
    cmd.add("git");
    for (String a : args) {
      cmd.add(a);
    }
    ProcessBuilder pb = new ProcessBuilder(cmd);
    pb.directory(dir.toFile());
    Process p = pb.start();
    boolean ok = p.waitFor(120, TimeUnit.SECONDS);
    if (!ok) {
      p.destroyForcibly();
      throw new RuntimeException("git timeout");
    }
    String stdout =
        new String(p.getInputStream().readAllBytes(), java.nio.charset.StandardCharsets.UTF_8);
    if (p.exitValue() != 0) {
      String err =
          new String(p.getErrorStream().readAllBytes(), java.nio.charset.StandardCharsets.UTF_8);
      throw new RuntimeException("git failed: " + err);
    }
    return stdout.trim();
  }

  @Test
  void sampleRepoFlowRuns() throws Exception {
    Path base = tmp;
    Path remote = base.resolve("origin.git");
    Path local = base.resolve("local");
    Files.createDirectories(remote);
    Files.createDirectories(local);

    runGit(remote, "init", "--bare");

    runGit(local, "init");
    runGit(local, "config", "user.email", "harness-sample@example.com");
    runGit(local, "config", "user.name", "git-harness sample");
    Files.writeString(local.resolve("README.md"), "hello\n");
    runGit(local, "add", "README.md");
    runGit(local, "commit", "-m", "init");

    CliBridge bridge = new CliBridge(workspaceRoot());
    String branch = bridge.getCurrentBranch(local.toString());

    runGit(local, "remote", "add", "origin", remote.toAbsolutePath().normalize().toString());
    runGit(local, "push", "-u", "origin", branch);

    String localSha = bridge.getCommitSHA(local.toString(), branch);
    String remoteSha = runGitStdout(remote, "rev-parse", branch);
    assertEquals(localSha, remoteSha, "SHA mismatch between local and remote");

    List<CliBridge.RepositoryMeta> repos =
        bridge.scanRepositories(
            new CliBridge.ScanOptions().rootPath(base.toString()).useCache(false).maxDepth(30));
    Path localReal = local.toRealPath();
    boolean found =
        repos.stream()
            .anyMatch(
                r -> {
                  try {
                    return Path.of(r.path()).toRealPath().equals(localReal);
                  } catch (IOException e) {
                    return false;
                  }
                });
    assertTrue(found, "scan_repositories did not find local repo");
  }
}
