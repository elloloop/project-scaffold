package com.elloloop.scaffold.shared;

import com.elloloop.scaffold.shared.serverkit.HealthResponse;
import com.elloloop.scaffold.shared.serverkit.HealthService;

public final class SharedApplication {
    private SharedApplication() {
    }

    public static void main(String[] args) {
        HealthResponse response = HealthService.health("shared_java");
        System.out.println(response.service() + " " + response.status());
    }
}
