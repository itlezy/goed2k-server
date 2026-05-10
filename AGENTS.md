# ED2K Server Repo Rules

- Follow the shared workspace policy in
  `../p2p-overlord-tooling/docs/WORKSPACE_POLICY.md`.
- Use `README.md` as the canonical ED2K server docs home.
- Use `../p2p-overlord-be/BACKLOG.md` as the canonical active backlog.
- This repo is the active local ED2K server used by Overlord parity scenarios.
- Keep the checkout at `%OVERLORD_PROJECT_DIR%\p2p-overlord-ed2k-server`;
  do not add a separate ED2K server path environment variable.
- Resolve the eMule harness only through `%EMULE_WORKSPACE_ROOT%`.
- Before finishing Go changes, run:
  - `go test ./...`
  - `go build -o %OVERLORD_TMP_DIR%\overlord-ed2k-server.exe .\cmd\overlord-ed2k-server`
- Keep tracked text files normalized to UTF-8 with LF endings; the workspace
  line-ending guard is enforced from tooling.
- Do not add shell wrapper launchers.