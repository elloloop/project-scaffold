package com.elloloop.scaffold.shared.serverkit;

import com.elloloop.scaffold.shared.platform.DisplayNames;

public final class HealthService {
    private HealthService() {
    }

    public static HealthResponse health(String service) {
        return new HealthResponse(DisplayNames.displayName(service), "ok");
    }
}
