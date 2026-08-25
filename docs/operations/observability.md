# Observability

Instrument important flows with OpenTelemetry-compatible traces, structured
logs, and metrics. Correlation fields: `request_id`, `trace_id`, `actor_id`,
`agent_id`, `team_id`, `repository_id`, `operation`.

Bootstrap currently exposes `/health` only. Broader instrumentation lands with
ingestion, retrieval, and MCP features.
