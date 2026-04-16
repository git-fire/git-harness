package io.gitfire.harness;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

class SampleSafetyFlowSmoke {
  @Test
  void sampleSafetyFlowRuns() {
    CliBridge bridge = new CliBridge(CliBridgeTest.workspaceRoot());
    String token = "ghp_" + "a".repeat(36);
    String out = bridge.safetySanitizeText("export TOKEN=" + token);
    assertFalse(out.contains(token));
    assertTrue(out.contains("[REDACTED]"));
    String notice = bridge.safetySecurityNotice();
    assertTrue(notice.length() > 10);
  }
}
