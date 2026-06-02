package com.elloloop.scaffold.shared.platform;

public final class DisplayNames {
    private DisplayNames() {
    }

    public static String displayName(String value) {
        String[] parts = value.split("[-_\\s]+");
        StringBuilder builder = new StringBuilder();

        for (String part : parts) {
            if (part.isBlank()) {
                continue;
            }

            if (!builder.isEmpty()) {
                builder.append(' ');
            }

            builder.append(Character.toUpperCase(part.charAt(0)));
            if (part.length() > 1) {
                builder.append(part.substring(1).toLowerCase());
            }
        }

        return builder.toString();
    }
}
