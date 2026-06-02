package com.elloloop.scaffold.shared.testkit;

import java.util.Objects;

public final class Assertions {
    private Assertions() {
    }

    public static <T> void assertEqual(T got, T want) {
        if (!Objects.equals(got, want)) {
            throw new AssertionError("got " + got + ", want " + want);
        }
    }
}
