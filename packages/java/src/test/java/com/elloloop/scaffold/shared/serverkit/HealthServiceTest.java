package com.elloloop.scaffold.shared.serverkit;

import com.elloloop.scaffold.shared.testkit.Assertions;
import org.junit.jupiter.api.Test;

class HealthServiceTest {
    @Test
    void returnsHealthResponse() {
        HealthResponse response = HealthService.health("project_scaffold");

        Assertions.assertEqual(response.service(), "Project Scaffold");
        Assertions.assertEqual(response.status(), "ok");
    }
}
