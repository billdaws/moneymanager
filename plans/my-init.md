I'm building an application that manages my finances. I'd like you to start building the skeleton of this project. 
Here are the high-level requirements:

- Written in Go
- Development tooling is managed via nix to the extent possible.
- Maintains any custom data it needs in SQLite
- Runs as a long-running service, not as a simple script.
- Accepts statements under an `/upload` API.
  - e.g. bank accounts, credit cards, etc.
- Uses kreuzberg to extract transaction from statements.
- Normalizes those transactions into a standard format.
- Loads those transactions into a GNU Cash database.
- Uses its own GNU Cash library, no third-party library for that.
- Dependencies are managed via Docker Compose.
  - I have provided an example compose stack that spins up Kreuzberg with no egress, to avoid telemetry.
- Caddy is used as a reverse proxy.

Ultimately I will be able to automatically forward/send it statements as they come to me so that they are
ingested into the GNU Cash database, and then I can use GNU Cash's reporting functionality on those transactions.

I'd like you to come up with a plan to implement these requirements.
You may track your progress in markdown files under a `task` directory.
