# Mobile Apps

This folder is a framework-neutral mobile surface. Do not preselect Flutter, React Native, Android, or iOS in the scaffold.

Supported setup options live under `options/`:

- `flutter`
- `react-native`
- `android`
- `ios`
- `shared`

When a project is instantiated, the agent should keep the chosen mobile stack, remove unused options, install the latest stable tooling, and preserve the relevant test categories under `tests/`.

Required mobile test categories:

- unit
- widget/component
- navigation
- integration
- end-to-end
- accessibility
- performance
- snapshot
- golden/visual
- contract/API client
- device/emulator
- localization
- permissions
- offline sync
- release smoke
