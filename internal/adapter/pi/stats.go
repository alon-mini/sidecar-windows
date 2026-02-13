package pi

// Pi sessions provide pre-calculated cost in usage.cost.total on each
// assistant message. Cost aggregation is done inline in adapter.go's
// sessionMetadata and Usage methods — no model-rate lookup is needed.
